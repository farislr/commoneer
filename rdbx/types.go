package rdbx

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
)

type contextKeyEnableSqlTx struct{}

// Assignabler
type Assignabler interface {
	Assign(value interface{}) (out interface{}, err error)
}

type DBTX interface {
	Queryx(ctx context.Context, q string, model interface{}, args ...interface{}) error

	dbsql
	reds
	UniversalTransactioner
}

type dbsql interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	Begin() (*sql.Tx, error)
	Close() error
}

type reds interface {
	redis.Cmdable
}

type UniversalTransactioner interface {
	TxEnd(ctx context.Context, txFunc interface{}, mutex *redsync.Mutex) error

	TxDefault
}

type TxDefault interface {
	EnableTx(ctx context.Context, txFunc TxFnNoMutex) error
	EnableTxWithRedisLock(ctx context.Context, key, value string, duration time.Duration, retVal interface{}, txFunc TxFn) error
}

type TxFn func(ctx context.Context, mutex *redsync.Mutex) error
type TxFnNoMutex func(ctx context.Context) error
