package rdbx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type txSuite struct {
	rdmock redismock.ClientMock
	mock   sqlmock.Sqlmock

	dbtx    DBTX
	rClient *redis.Client
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(txSuite))
}

func (t *txSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.NoError(err)
	}
	t.mock = mock

	rdclient, rdmock := redismock.NewClientMock()
	t.rdmock = rdmock

	t.rClient = rdclient
	t.dbtx = NewDbx(db, &cache{redis: rdclient})
}

func (t *txSuite) Test_tx_Lock() {
	txx := NewTransactioner(t.dbtx, t.rClient)

	t.mock.ExpectBegin()
	t.mock.ExpectBegin()

	t.mock.ExpectCommit()
	t.mock.ExpectBegin()

	t.mock.ExpectCommit()
	// t.mock.ExpectCommit()

	t.rdmock.Regexp().ExpectSetNX("key-1", ``, 30*time.Second).SetVal(true)
	t.rdmock.ExpectSetNX("key-1.value", `321321`, 30*time.Second).SetVal(true)

	t.rdmock.Regexp().ExpectSetNX("key-1", ``, 30*time.Second).SetVal(false)
	t.rdmock.Regexp().
		ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-1"}, ``).
		SetVal(true)

		// t.rdmock.ExpectSetNX("key-1.value", "321321321", 30*time.Second).SetVal(false)
		// t.rdmock.ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-1.value"}, "321321321").SetVal(true)

	// t.rdmock.ExpectDel("key-1.value")
	t.rdmock.Regexp().ExpectSetNX("key-2", ``, 30*time.Second).SetVal(true)
	t.rdmock.ExpectSetNX("key-2.value", "321321", 30*time.Second).SetVal(true)
	t.rdmock.Regexp().
		ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-2"}, ``).
		SetVal(true)
	t.rdmock.ExpectDel("key-2.value").SetVal(1)
	// t.rdmock.ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-1"}, "321321321").SetVal(false)
	// t.rdmock.ExpectGet("123123").SetVal("321321")

	ctx := context.Background()

	tx1 := txx.EnableTx(ctx).
		WithRedisLock("key-1").
		WithLockDuration(30 * time.Second).
		WithValue("321321")

	tx2 := txx.EnableTx(ctx).
		WithRedisLock("key-2", AutoUnlock()).
		WithLockDuration(30 * time.Second).
		WithValue("321321")

	err1 := tx1.Exec(func(ctx context.Context, m Redsync) error {
		return nil
	})
	fmt.Printf("err1: %v\n", err1)

	tx3 := txx.EnableTx(ctx).
		WithRedisLock("key-1").
		WithLockDuration(30 * time.Second).
		WithValue("321321321")

	err3 := tx3.Exec(func(ctx context.Context, mutex Redsync) error {
		// ok, err := mutex.ValidContext(ctx)
		// fmt.Printf("ok: %v\n", ok)
		// fmt.Printf("err: %v\n", err)
		return nil
	})
	fmt.Printf("err3: %v\n", err3)

	err2 := tx2.Exec(func(ctx context.Context, m Redsync) error {
		return nil
	})
	fmt.Printf("err2: %v\n", err2)

	assert.NoError(t.T(), t.mock.ExpectationsWereMet())
	assert.NoError(t.T(), t.rdmock.ExpectationsWereMet())

	// ok, errx := tx1.Unlock(ctx)
	// ok, errx := tx2.Unlock(ctx)

	// fmt.Printf("ok: %v\n", ok)
	// fmt.Printf("errx: %v\n", errx)
}

func (t *txSuite) Test_tx_GetLock() {
	txx := NewTransactioner(t.dbtx, t.rClient)

	t.mock.ExpectBegin()
	t.mock.ExpectCommit()

	t.rdmock.Regexp().ExpectSetNX("key-1", ``, 30*time.Second).SetVal(true)
	t.rdmock.ExpectSetNX("key-1.value", "123123", 30*time.Second).SetVal(true)
	t.rdmock.ExpectGet("key-1.value").SetVal("123123")
	t.rdmock.ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-1"}, ``).
		SetVal(int64(1))
	t.rdmock.ExpectDel("key-1.value").SetVal(1)
	t.rdmock.ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-2"}, "").
		SetVal(int64(0))
	t.rdmock.ExpectGet("key-2").SetVal("123123")
	t.rdmock.ExpectEvalSha("00583931c4e4d483f6233879c133c71ed5393f9d", []string{"key-2"}, "123123").
		SetVal(int64(1))
	t.rdmock.ExpectDel("key-2.value").SetVal(1)

	ctx := context.Background()

	err := txx.EnableTx(ctx).
		WithRedisLock("key-1").
		WithLockDuration(30 * time.Second).
		WithValue("123123").
		Exec(func(ctx context.Context, rtx Redsync) error {
			return nil
		})

	fmt.Printf("err: %v\n", err)

	tx1 := txx.EnableTx(ctx).
		WithRedisLock("key-1")

	val, err := tx1.GetValue(ctx)
	fmt.Printf("val: %v\n", val)
	fmt.Printf("err: %v\n", err)

	ok, err := tx1.Unlock()

	fmt.Printf("ok: %v\n", ok)
	fmt.Printf("err: %v\n", err)

	tx2 := txx.EnableTx(ctx).WithRedisLock("key-2")

	fmt.Println(tx2.Unlock())

}
