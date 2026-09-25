package cache

import (
	"context"
	"time"
)

type Entry struct {
	Data             []byte
	HasData          bool
	UpdatedAt        time.Time
	RefreshStartedAt *time.Time
	LastRefreshError *string
}

type Store interface {
	Get(ctx context.Context, key string) (*Entry, error)
	TryStartRefresh(ctx context.Context, key string, lease time.Duration) (bool, error)
	Save(ctx context.Context, key string, data []byte) error
	FailRefresh(ctx context.Context, key string, err error) error
}
