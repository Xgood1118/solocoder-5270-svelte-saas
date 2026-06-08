package db

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type contextKey string

const orgIDKey contextKey = "org_id"

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	dsn := dbPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(time.Hour)

	if err := database.Ping(); err != nil {
		return nil, err
	}

	return &DB{database}, nil
}

func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDKey, orgID)
}

func OrgIDFromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(orgIDKey).(string)
	return orgID, ok
}

func (db *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
