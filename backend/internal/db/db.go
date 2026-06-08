package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type contextKey string

const orgIDKey contextKey = "org_id"

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	database, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
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
