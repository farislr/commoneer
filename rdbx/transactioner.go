package rdbx

import (
	"context"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
)

type tx struct {
	DBTX
}

func NewTransactioner(db DBTX) UniversalTransactioner {
	return &tx{
		db,
	}
}

func (dbx *dbx) TxEnd(ctx context.Context, txFunc interface{}, mutex *redsync.Mutex) error {
	var err error
	tx, err := dbx.Begin()
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, &contextKeyEnableSqlTx{}, tx)

	defer func() {
		p := recover()

		switch {
		case p != nil:

			if err = tx.Rollback(); err != nil {
				panic(err)
			}

		case err != nil:
			if err = tx.Rollback(); err != nil {
				log.Println(err)
			}

			dbx.unlock(ctx, mutex)
		default:
			err = tx.Commit()
			if err != nil {
				log.Println(err)
			}
		}
	}()

	switch fun := txFunc.(type) {
	case TxFn:
		err = fun(ctx, mutex)
	case TxFnNoMutex:
		err = fun(ctx)
	}

	return err
}

func (dbx *dbx) EnableTx(ctx context.Context, txFunc TxFnNoMutex) error {
	return dbx.TxEnd(ctx, txFunc, nil)
}

func (dbx *dbx) EnableTxWithRedisLock(ctx context.Context, key, value string, duration time.Duration, retVal interface{}, txFunc TxFn) error {
	mutex := dbx.NewMutex(key, redsync.WithExpiry(duration), redsync.WithTries(1), redsync.WithGenValueFunc(func() (string, error) { return value, nil }))

	if err := mutex.LockContext(ctx); err != nil {
		if err := dbx.Get(ctx, key).Scan(retVal); err != nil {
			return err
		}

		return err
	}

	return dbx.TxEnd(ctx, txFunc, mutex)
}

func (dbx *dbx) unlock(ctx context.Context, mutex *redsync.Mutex) {
	if mutex != nil {
		if ok, err := mutex.UnlockContext(ctx); !ok || err != nil {
			log.Println(err)
		}
	}
}
