package rdbx

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	Begin() (*sql.Tx, error)
	Close() error
}

type Transactioner interface {
	EnableTx(ctx context.Context) *enabledTx
}

type Redsync interface {
	GetValue(ctx context.Context) (string, error)
	Unlock() (bool, error)
}
