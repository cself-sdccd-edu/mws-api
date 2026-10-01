package cache

import (
	"context"
	"fmt"
	mwslog "github.com/cself-sdccd-edu/mws-api/internal/log"
	"github.com/cself-sdccd-edu/mws-api/internal/qas"
	"log"
	"time"
)

type Service struct {
	store          Store
	qasClient      qas.Client
	logger         *log.Logger
	eventLogger    mwslog.Logger
	cacheTime      time.Duration
	maxCacheTime   time.Duration
	refreshLease   time.Duration
	refreshTimeout time.Duration
}

type RefreshRequest struct {
	EndpointName string
	QueryName    string
	Term         string
}

func NewService(store Store,
	qasClient qas.Client,
	logger *log.Logger,
	eventLogger mwslog.Logger,
	cacheTime time.Duration,
	maxCacheTime time.Duration,
	refreshLease time.Duration,
	refreshTimeout time.Duration) *Service {

	return &Service{
		store:          store,
		maxCacheTime:   maxCacheTime,
		cacheTime:      cacheTime,
		refreshLease:   refreshLease,
		refreshTimeout: refreshTimeout,
		qasClient:      qasClient,
		logger:         logger,
		eventLogger:    eventLogger,
	}
}

func (s *Service) Get(ctx context.Context, key string, refresh RefreshRequest) (*Entry, error) {
	entry, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	requestID := mwslog.RequestID(ctx)

	if entry != nil && entry.HasData {
		age := time.Since(entry.UpdatedAt)
		s.logger.Printf("requestid %v cache %q age=%v cache_time=%v max_cache_time=%v updated_at=%v", requestID, key, age, s.cacheTime, s.maxCacheTime, entry.UpdatedAt)
		if age < s.cacheTime {
			s.logEvent(ctx, mwslog.LogEvent{
				RequestID: requestID,
				Event:     "cache_hit",
				CacheKey:  key,
				Message:   "cache entry is fresh",
			})

			return entry, nil
		}

		if age < s.maxCacheTime {
			s.logEvent(ctx, mwslog.LogEvent{
				RequestID: requestID,
				Event:     "cache_hit",
				CacheKey:  key,
				Message:   "cache entry is stale but within maximum cache lifetime",
			})

			claimed, err := s.store.TryStartRefresh(ctx, key, s.refreshLease)
			if err != nil {
				s.logEvent(ctx, mwslog.LogEvent{
					RequestID: requestID,
					Event:     "cache_refresh_claim_error",
					CacheKey:  key,
					QueryName: refresh.QueryName,
					Term:      refresh.Term,
					Message:   err.Error(),
				})

				return entry, nil
			}

			if claimed {
				s.logEvent(ctx, mwslog.LogEvent{
					RequestID: requestID,
					Event:     "cache_refresh_start",
					CacheKey:  key,
					QueryName: refresh.QueryName,
					Term:      refresh.Term,
					Message:   "background cache refresh started",
				})

				refreshCtx := mwslog.WithRequestID(context.Background(), requestID)
				go s.refresh(refreshCtx, key, refresh)
			}

			return entry, nil
		}
	}

	s.logEvent(ctx, mwslog.LogEvent{
		RequestID: requestID,
		Event:     "cache_miss",
		CacheKey:  key,
		QueryName: refresh.QueryName,
		Term:      refresh.Term,
		Message:   "cache entry is missing or expired",
	})

	claimed, err := s.store.TryStartRefresh(ctx, key, s.refreshLease)
	if err != nil {
		return nil, fmt.Errorf("claim cache refresh: %w", err)
	}

	if !claimed {
		return s.waitForRefresh(ctx, key, refresh)
	}

	return s.refreshNow(ctx, key, refresh)

}

func (s *Service) refreshNow(ctx context.Context, key string, refresh RefreshRequest) (*Entry, error) {
	start := time.Now()
	requestID := mwslog.RequestID(ctx)

	data, err := s.qasClient.Query(ctx, qas.QueryRequest{
		QueryName: refresh.QueryName,
		Term:      refresh.Term,
	})
	duration := time.Since(start)

	if err != nil {
		s.logEvent(ctx, mwslog.LogEvent{
			RequestID:  requestID,
			Event:      "qas_error",
			CacheKey:   key,
			QueryName:  refresh.QueryName,
			Term:       refresh.Term,
			DurationMs: int(duration.Milliseconds()),
			Message:    err.Error(),
		})

		s.failRefresh(key, err)

		return nil, fmt.Errorf("QAS refresh failed: %w", err)
	}

	s.logEvent(ctx, mwslog.LogEvent{
		RequestID:  requestID,
		Event:      "qas_refresh",
		CacheKey:   key,
		QueryName:  refresh.QueryName,
		Term:       refresh.Term,
		DurationMs: int(duration.Milliseconds()),
		DataSize:   len(data),
		Message:    "QAS refresh completed",
	})

	if err := s.store.Save(ctx, key, data); err != nil {
		s.logEvent(ctx, mwslog.LogEvent{
			RequestID: requestID,
			Event:     "cache_save_error",
			CacheKey:  key,
			QueryName: refresh.QueryName,
			Term:      refresh.Term,
			Message:   err.Error(),
		})

		s.failRefresh(key, err)
		//if failErr := s.store.FailRefresh(ctx, key, err); failErr != nil {
		//	return nil, fmt.Errorf("save refreshed cache: %w; recording failure: %v", err, failErr)
		//}

		return nil, fmt.Errorf("save refreshed cache: %w", err)
	}

	entry, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get refreshed cache: %w", err)
	}

	return entry, nil

}

func (s *Service) refresh(ctx context.Context, key string, refresh RefreshRequest) {
	ctx, cancel := context.WithTimeout(ctx, s.refreshTimeout)
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
			s.logEvent(ctx, mwslog.LogEvent{
				RequestID: mwslog.RequestID(ctx),
				Event:     "cache_refresh_complete",
				CacheKey:  key,
				Message:   "waited for another request to refresh cache",
			})

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

func (s *Service) failRefresh(key string, err error) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if failErr := s.store.FailRefresh(cleanupCtx, key, err); failErr != nil {
		s.logger.Printf("cache %q: failed to clear refresh lease after error %v: %v", key, err, failErr)
	}
}

func (s *Service) logEvent(ctx context.Context, event mwslog.LogEvent) {
	if err := s.eventLogger.Log(ctx, event); err != nil {
		s.logger.Printf("event logging failed: %v", err)
	}
}
