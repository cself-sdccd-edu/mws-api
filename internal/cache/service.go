package cache

import (
	"context"
	"fmt"
	"github.com/cself-sdccd-edu/mws-api/internal/qas"
	"log"
	"time"
)

type Service struct {
	store        Store
	qasClient    qas.Client
	logger       *log.Logger
	cacheTime    time.Duration
	maxCacheTime time.Duration
	refreshLease time.Duration
}

type RefreshRequest struct {
	QueryName string
	Term      string
}

func NewService(store Store, qasClient qas.Client, logger *log.Logger, cacheTime time.Duration, maxCacheTime time.Duration, refreshLease time.Duration) *Service {
	return &Service{
		store:        store,
		cacheTime:    cacheTime,
		maxCacheTime: maxCacheTime,
		refreshLease: refreshLease,
		qasClient:    qasClient,
		logger:       logger,
	}
}

func (s *Service) Get(ctx context.Context, key string, refresh RefreshRequest) (*Entry, error) {
	entry, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if entry != nil && entry.HasData {
		age := time.Since(entry.UpdatedAt)
		s.logger.Printf("cache %q age=%v cache_time=%v max_cache_time=%v updated_at=%v", key, age, s.cacheTime, s.maxCacheTime, entry.UpdatedAt)

		// best case, we have valid cached data - let's just return it
		if age < s.cacheTime {
			return entry, nil
		}

		// we have stale data that is still within the maximum cache lifetime.
		// fire up a refresh background job to update cache but return the stale data.
		if age < s.maxCacheTime {
			claimed, err := s.store.TryStartRefresh(ctx, key, s.refreshLease)
			if err != nil {
				return nil, fmt.Errorf("claim cache refresh: %w", err)
			}

			if claimed {
				go s.refresh(key, refresh)
			}

			return entry, nil
		}
	}

	// there is no usable data, try to claim a refresh lease
	claimed, err := s.store.TryStartRefresh(ctx, key, s.refreshLease)
	if err != nil {
		return nil, fmt.Errorf("claim cache refresh: %w", err)
	}

	// someone else is refreshing so let's wait. waiting will also attempt to claim
	// a lease after some time in order to recover if the other refresh task fails.
	if !claimed {
		return s.waitForRefresh(ctx, key, refresh)
	}

	// there is no usable data to serve, so refresh it here, in this context, and wait.
	return s.refreshNow(ctx, key, refresh)
}

func (s *Service) refreshNow(ctx context.Context, key string, refresh RefreshRequest) (*Entry, error) {
	data, err := s.qasClient.Query(ctx, refresh.QueryName, refresh.Term)
	if err != nil {
		if failErr := s.store.FailRefresh(ctx, key, err); failErr != nil {
			return nil, fmt.Errorf("QAS refresh failed: %w; recording failure: %v", err, failErr)
		}

		return nil, fmt.Errorf("QAS refresh failed: %w", err)
	}

	if err := s.store.Save(ctx, key, data); err != nil {
		if failErr := s.store.FailRefresh(ctx, key, err); failErr != nil {
			return nil, fmt.Errorf("save refreshed cache: %w; recording failure: %v", err, failErr)
		}

		return nil, fmt.Errorf("save refreshed cache: %w", err)
	}
	s.logger.Printf("returned from QAS fetch for %s", key)
	entry, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get refreshed cache: %w", err)
	}

	return entry, nil
}

func (s *Service) refresh(key string, refresh RefreshRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), s.refreshLease)
	defer cancel()

	if _, err := s.refreshNow(ctx, key, refresh); err != nil {
		s.logger.Printf("background cache refresh %q failed: %v", key, err)
	}
}

func (s *Service) waitForRefresh(ctx context.Context, key string, refresh RefreshRequest) (*Entry, error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}

		entry, err := s.store.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		if entry != nil && entry.HasData && time.Since(entry.UpdatedAt) < s.maxCacheTime {
			return entry, nil
		}

		claimed, err := s.store.TryStartRefresh(ctx, key, s.refreshLease)
		if err != nil {
			return nil, fmt.Errorf("claim cache refresh while waiting: %w", err)
		}

		if claimed {
			return s.refreshNow(ctx, key, refresh)
		}
	}
}
