package log

import (
	"context"
)

type Logger interface {
	Log(ctx context.Context, event LogEvent) error
}

type LogEvent struct {
	RequestID  string
	Event      string
	ClientIP   string
	UserAgent  string
	Method     string
	Endpoint   string
	StatusCode int
	DurationMs int
	CacheKey   string
	QueryName  string
	Term       string
	DataSize   int
	Message    string
	Details    string
}
