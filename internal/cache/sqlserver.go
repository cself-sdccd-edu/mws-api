package cache

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLServerStore struct {
	db *sql.DB
}

func NewSQLServerStore(db *sql.DB) *SQLServerStore {
	return &SQLServerStore{db: db}
}

func (s *SQLServerStore) Get(ctx context.Context, key string) (*Entry, error) {
	row := s.db.QueryRowContext(ctx, "EXEC dbo.Cache_Get @CacheKey = @p1", sql.Named("p1", key))

	var entry Entry
	var refreshStartedAt sql.NullTime
	var lastRefreshError sql.NullString

	err := row.Scan(&key, &entry.Data, &entry.UpdatedAt, &refreshStartedAt, &lastRefreshError)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get cache entry %q: %w", key, err)
	}

	if refreshStartedAt.Valid {
		entry.RefreshStartedAt = &refreshStartedAt.Time
	}

	if lastRefreshError.Valid {
		entry.LastRefreshError = &lastRefreshError.String
	}

	return &entry, nil
}
func (s *SQLServerStore) TryStartRefresh(ctx context.Context, key string, lease time.Duration) (bool, error) {
	row := s.db.QueryRowContext(ctx, "EXEC dbo.Cache_TryStartRefresh @CacheKey = @p1, @LeaseSeconds = @p2", sql.Named("p1", key), sql.Named("p2", int(lease.Seconds())))

	var claimed bool
	if err := row.Scan(&claimed); err != nil {
		return false, fmt.Errorf("try start cache refresh %q: %w", key, err)
	}

	return claimed, nil
}
func (s *SQLServerStore) Save(ctx context.Context, key string, data []byte) error {
	_, err := s.db.ExecContext(ctx, "EXEC dbo.Cache_Save @CacheKey = @p1, @Data = @p2", sql.Named("p1", key), sql.Named("p2", data))
	if err != nil {
		return fmt.Errorf("save cache entry %q: %w", key, err)
	}

	return nil
}
func (s *SQLServerStore) FailRefresh(ctx context.Context, key string, refreshErr error) error {
	_, err := s.db.ExecContext(ctx, "EXEC dbo.Cache_FailRefresh @CacheKey = @p1, @Error = @p2", sql.Named("p1", key), sql.Named("p2", refreshErr.Error()))
	if err != nil {
		return fmt.Errorf("fail cache refresh %q: %w", key, err)
	}

	return nil
}
