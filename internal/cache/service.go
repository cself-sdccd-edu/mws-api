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
	qasClient    *qas.HTTPClient
	logger       *log.Logger
	cacheTime    time.Duration
	refreshLease time.Duration
}

type Result struct {
	Entry         *Entry
	RefreshNeeded bool
}

type RefreshRequest struct {
	QueryName string
	Term      string
}

func NewService(store Store, qasClient *qas.HTTPClient, logger *log.Logger, cacheTime time.Duration, refreshLease time.Duration) *Service {
	return &Service{
		store:        store,
		cacheTime:    cacheTime,
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

	// best case, we have valid cached data - let's just return it
	if entry != nil && entry.Data != nil && time.Since(entry.UpdatedAt) < s.cacheTime {
		return entry, nil
	}

	// no valid data, try to claim a refresh lease
	claimed, err := s.store.TryStartRefresh(ctx, key, s.refreshLease)
	if err != nil {
		return nil, fmt.Errorf("claim cache refresh: %w", err)
	}

	// we have some existing data that we can immediately save, but it's stale.
	// fire up a refresh background job to update cache but return the stale data.
	if entry != nil && entry.Data != nil {
		if claimed {
			go s.refresh(key, refresh)
		}

		return entry, nil
	}

	// someone else is refreshing so let's wait. waiting will also attempt to claim
	// a lease after some time in order to recover if the other refresh task fails.
	if !claimed {
		return s.waitForRefresh(ctx, key)
	}

	// there is no stale data to serve, so refresh it here, in this context, and wait.
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
		return nil, fmt.Errorf("save refreshed cache: %w", err)
	}

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

func (s *Service) waitForRefresh(ctx context.Context, key string) (*Entry, error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		entry, err := s.store.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		if entry != nil && entry.Data != nil {
			return entry, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
