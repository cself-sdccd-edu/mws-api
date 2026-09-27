package log

import (
	"context"
	"database/sql"
)

type SQLLogger struct {
	db           *sql.DB
	serverNumber int
	version      string
}

func NewSQLLogger(db *sql.DB, serverNumber int, version string) *SQLLogger {
	return &SQLLogger{db: db, serverNumber: serverNumber, version: version}
}

func (l *SQLLogger) Log(ctx context.Context, event LogEvent) error {
	_, err := l.db.ExecContext(ctx, "dbo.Log_Insert",
		sql.Named("RequestID", nullString(event.RequestID)),
		sql.Named("Event", event.Event),
		sql.Named("ClientIP", nullString(event.ClientIP)),
		sql.Named("UserAgent", nullString(event.UserAgent)),
		sql.Named("ServerNumber", l.serverNumber),
		sql.Named("Version", nullString(l.version)),
		sql.Named("Method", nullString(event.Method)),
		sql.Named("Endpoint", nullString(event.Endpoint)),
		sql.Named("StatusCode", nullInt(event.StatusCode)),
		sql.Named("DurationMs", nullInt(event.DurationMs)),
		sql.Named("CacheKey", nullString(event.CacheKey)),
		sql.Named("QueryName", nullString(event.QueryName)),
		sql.Named("Term", nullString(event.Term)),
		sql.Named("DataSize", nullInt(event.DataSize)),
		sql.Named("Message", nullString(event.Message)),
		sql.Named("Details", nullString(event.Details)),
	)
	return err
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
