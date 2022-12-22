package rdbx

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
)

type enabledRedisTx struct {
	tx *enabledTx

	value        string
	lockDuration time.Duration
}

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
		value:        "",
		lockDuration: 180 * time.Second,
	}
}

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

func (rtx *enabledRedisTx) WithValue(value string) *enabledRedisTx {
	rtx.value = value

	return rtx
}

func (rtx *enabledRedisTx) WithLockDuration(duration time.Duration) *enabledRedisTx {
	rtx.lockDuration = duration

	rtx.tx.m = rtx.tx.rsync.NewMutex(
		rtx.tx.key,
		redsync.WithTries(1),
		redsync.WithExpiry(duration),
	)

	return rtx
}

func (rtx *enabledRedisTx) GetValue(ctx context.Context) (string, error) {
	res := rtx.tx.redisClient.
		Get(ctx, fmt.Sprintf("%v.value", rtx.tx.key))

	return res.Result()
}

func (rtx *enabledRedisTx) lock(ctx context.Context) error {
	if err := rtx.tx.m.LockContext(ctx); err != nil {
		return err
	}

	if rtx.value != "" {
		if err := rtx.tx.redisClient.SetNX(ctx, fmt.Sprintf("%v.value", rtx.tx.key), rtx.value, rtx.lockDuration).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (rtx *enabledRedisTx) Lock() error {
	ctx := rtx.tx.ctx

	return rtx.lock(ctx)
}

func (rtx *enabledRedisTx) Unlock() (bool, error) {
	ctx := rtx.tx.ctx

	return rtx.unlock(ctx)
}

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

	if err := rtx.tx.redisClient.Del(ctx, fmt.Sprintf("%v.value", rtx.tx.key)).Err(); err != nil {
		return false, err
	}

	return true, nil
}

type RedisLockOption interface {
	Apply(*enabledTx)
}

type optionFunc func(*enabledTx)

func (f optionFunc) Apply(t *enabledTx) {
	f(t)
}

func AutoUnlock() RedisLockOption {
	return optionFunc(func(t *enabledTx) {
		t.autoUnlock = true
	})
}
