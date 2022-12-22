// Package rdbx provides a Redis-based transaction manager with distributed locking capabilities.
// This file contains the implementation of the Redis-based distributed locking mechanism.
package rdbx

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
)

// enabledRedisTx represents a Redis-based transaction with distributed locking capabilities.
type enabledRedisTx struct {
	tx *enabledTx

	lockDuration time.Duration
}

// WithRedisLock creates a new Redis-based transaction with distributed locking capabilities.
func (t *enabledTx) WithRedisLock(key string, options ...RedisLockOption) *enabledRedisTx {
	for _, o := range options {
		o.Apply(t)
	}

	t.key = key

	t.m = t.rsync.NewMutex(
		t.key,
		redsync.WithTries(1),
	)

	return &enabledRedisTx{
		tx:           t,
		lockDuration: 180 * time.Second,
	}
}

// Exec executes the Redis-based transaction with distributed locking capabilities.
func (rtx *enabledRedisTx) Exec(actionFn func(ctx context.Context, rtx Redsync) error) error {
	if rtx.tx.err = rtx.lock(rtx.tx.ctx); rtx.tx.err != nil {
		return rtx.tx.err
	}

	defer func() {
		if rtx.tx.m != nil {
			if rtx.tx.autoUnlock || rtx.tx.err != nil {
				if ok, err := rtx.unlock(rtx.tx.ctx); err != nil {
					log.Printf("[Transactioner Error Unlock] %v:%v", err, ok)
				}
			}
		}
	}()

	return rtx.tx.Exec(func(ctx context.Context) error {
		return actionFn(ctx, rtx)
	})
}

// WithLockDuration sets the lock duration of the Redis-based transaction.
func (rtx *enabledRedisTx) WithLockDuration(duration time.Duration) *enabledRedisTx {
	rtx.lockDuration = duration

	rtx.tx.m = rtx.tx.rsync.NewMutex(
		rtx.tx.key,
		redsync.WithTries(1),
		redsync.WithExpiry(duration),
	)

	return rtx
}

// lock acquires the lock for the Redis-based transaction.
func (rtx *enabledRedisTx) lock(ctx context.Context) error {
	if err := rtx.tx.m.LockContext(ctx); err != nil {
		return err
	}

	return nil
}

// Lock acquires the lock for the Redis-based transaction.
func (rtx *enabledRedisTx) Lock() error {
	ctx := rtx.tx.ctx

	return rtx.lock(ctx)
}

// Unlock releases the lock for the Redis-based transaction.
func (rtx *enabledRedisTx) Unlock() (bool, error) {
	ctx := rtx.tx.ctx

	return rtx.unlock(ctx)
}

// unlock releases the lock for the Redis-based transaction.
func (rtx *enabledRedisTx) unlock(ctx context.Context) (bool, error) {
	if rtx.tx.m == nil {
		return false, errors.New("no lock found")
	}

TryUnlock:
	ok, err := rtx.tx.m.UnlockContext(ctx)
	if err != nil {
		return false, err
	}

	if !ok {
		lockValue, err := rtx.tx.redisClient.Get(ctx, rtx.tx.key).Result()
		if err != nil {
			return false, err
		}

		rtx.tx.m = rtx.tx.rsync.NewMutex(
			rtx.tx.key,
			redsync.WithTries(1),
			redsync.WithExpiry(rtx.lockDuration),
			redsync.WithValue(lockValue),
		)

		goto TryUnlock
	}

	return true, nil
}

// RedisLockOption represents an option for the Redis-based transaction with distributed locking capabilities.
type RedisLockOption interface {
	Apply(*enabledTx)
}

// optionFunc represents a function that applies an option to the Redis-based transaction with distributed locking capabilities.
type optionFunc func(*enabledTx)

// Apply applies the option to the Redis-based transaction with distributed locking capabilities.
func (f optionFunc) Apply(t *enabledTx) {
	f(t)
}

// AutoUnlock returns an option that enables automatic unlocking of the Redis-based transaction with distributed locking capabilities.
func AutoUnlock() RedisLockOption {
	return optionFunc(func(t *enabledTx) {
		t.autoUnlock = true
	})
}
