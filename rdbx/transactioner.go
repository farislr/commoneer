package rdbx

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

type tx struct {
	dbtx        DBTX
	rsync       *redsync.Redsync
	redisClient redis.UniversalClient
}

func NewTransactioner(dbtx DBTX, redisCLient redis.UniversalClient) *tx {
	pool := goredis.NewPool(redisCLient)
	rsync := redsync.New(pool)

	return &tx{
		dbtx:        dbtx,
		rsync:       rsync,
		redisClient: redisCLient,
	}
}

func (t *tx) EnableTx(ctx context.Context) *enabledTx {
	tx, err := t.dbtx.Begin()

	ctx = context.WithValue(ctx, &contextKeyEnableSqlTx{}, tx)

	return &enabledTx{
		ctx:         ctx,
		rsync:       t.rsync,
		redisClient: t.redisClient,

		err: err,
	}
}

// enabledTx represents an enabled transaction with a Redis database.
type enabledTx struct {
	ctx         context.Context       // The context of the transaction.
	rsync       *redsync.Redsync      // The Redsync instance used for distributed locking.
	m           *redsync.Mutex        // The mutex used for locking.
	redisClient redis.UniversalClient // The Redis client used for the transaction.

	key string // The key of the transaction.
	err error  // The error of the transaction.

	autoUnlock bool // Whether the transaction should be automatically unlocked.
}

// Exec executes the given action function within the transaction context.
// If an error occurred during the transaction, it will be rolled back.
// If the action function panics, the transaction will be rolled back.
// If the action function returns an error, the transaction will be rolled back.
// Otherwise, the transaction will be committed.
func (t enabledTx) Exec(actionFn func(ctx context.Context) error) error {
	if t.err != nil {
		return t.err
	}

	tx, ok := t.ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx)
	if !ok {
		return errors.New("no tx found")
	}

	defer func(t *enabledTx) {
		r := recover()
		if tx != nil {
			switch {
			case r != nil:
				if t.err = tx.Rollback(); t.err != nil {
					log.Panicf("[Transactioner Error Rollback] %v", t.err)
				}
			case t.err != nil:
				if t.err = tx.Rollback(); t.err != nil {
					log.Printf("[Transactioner Error Rollback] %v", t.err)
				}
			default:
				if t.err = tx.Commit(); t.err != nil {
					log.Printf("[Transactioner Error Commit] %v", t.err)
				}
			}
		}

	}(&t)

	t.err = actionFn(t.ctx)

	return t.err
}
