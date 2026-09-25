package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

func Open(ctx context.Context, server string, database string, user string, password string, trustcert bool) (*sql.DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=true;TrustServerCertificate=%t", server, user, password, database, trustcert)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("open SQL Server connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping SQL Server: %w", err)
	}

	return db, nil
}
