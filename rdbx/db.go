package rdbx

import (
	"context"
	"database/sql"
	"errors"
	"reflect"

	"github.com/farislr/commoneer/rdbx/internal"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type dbx struct {
	DB *sql.DB

	redis.Cmdable
	*redsync.Redsync
}

func NewDBx(conn *sql.DB, r redis.Cmdable, rsync *redsync.Redsync) DBTX {
	return &dbx{
		conn,
		r,
		rsync,
	}
}

func (db *dbx) Queryx(ctx context.Context, query string, model interface{}, args ...interface{}) error {
	p := reflect.Indirect(reflect.ValueOf(model))

	t := p.Type()

	if (t.Kind() != reflect.Struct || t.Kind() != reflect.Slice) && !p.CanAddr() {
		return errors.New("model should be pointer, pointer struct, or pointer slice struct")
	}

	query = internal.ModifyOrKeepField(query, model)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {

		return err
	}
	defer rows.Close()

	err = db.rowScan(ctx, rows, p, t)
	if err != nil {
		return err
	}

	return nil
}

func (db *dbx) rowScan(ctx context.Context, rows *sql.Rows, rVal reflect.Value, rType reflect.Type) error {
	el := rVal

	if rVal.Kind() == reflect.Slice {
		el = reflect.Indirect(reflect.New(rVal.Type().Elem()))
		rType = el.Type()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	count := len(columns)
	vs := make([]interface{}, count)
	vPtrs := make([]interface{}, count)

	for rows.Next() {
		for i := range columns {
			vPtrs[i] = &vs[i]
		}

		err := rows.Scan(vPtrs...)
		if err != nil {
			return err
		}

		for i := 0; i < rType.NumField(); i++ {
			if c, ok := rType.Field(i).Tag.Lookup("column"); ok {
				for ii, col := range columns {
					if col == c {
						colVal := vs[ii]

						if b, ok := colVal.([]byte); ok {
							colVal = string(b)
						}

						if err = db.structFieldSet(el.Field(i), colVal); err != nil {
							return err
						}
					}
				}
			}
		}

		if rVal.Kind() == reflect.Slice {
			rVal.Set(reflect.Append(rVal, el))
		}
	}

	return nil
}

func (db *dbx) structFieldSet(element reflect.Value, value interface{}) error {
	if element.IsValid() {
		if element.CanSet() {
			switch v := element.Interface().(type) {
			case Assignabler:
				a, err := v.Assign(value)
				if err != nil {
					return err
				}
				element.Set(reflect.ValueOf(a))
			case sql.NullTime:
				if err := v.Scan(value); err != nil {
					return err
				}
				element.Set(reflect.ValueOf(v))
			case sql.NullString:
				if err := v.Scan(value); err != nil {
					return err
				}
				element.Set(reflect.ValueOf(v))
			case sql.NullInt64:
				if err := v.Scan(value); err != nil {
					return err
				}
				element.Set(reflect.ValueOf(v))
			case uuid.UUID:
				id, err := uuid.Parse(value.(string))
				if err != nil {
					return err
				}
				element.Set(reflect.ValueOf(id))
			case decimal.Decimal:
				v, err := decimal.NewFromString(value.(string))
				if err != nil {
					return err
				}
				element.Set(reflect.ValueOf(v))
			default:
				element.Set(reflect.ValueOf(value))
			}
		}
	}

	return nil
}

func (dbx *dbx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return dbx.DB.Exec(query, args...)
}

func (dbx *dbx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx)
	if ok {
		return tx.ExecContext(ctx, query, args...)
	}

	return dbx.DB.ExecContext(ctx, query, args...)
}

func (dbx *dbx) Prepare(query string) (*sql.Stmt, error) {
	return dbx.DB.Prepare(query)
}

func (dbx *dbx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx)
	if ok {
		return tx.PrepareContext(ctx, query)
	}

	return dbx.DB.PrepareContext(ctx, query)
}

func (dbx *dbx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return dbx.DB.Query(query, args...)
}

func (dbx *dbx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx)
	if ok {
		return tx.QueryContext(ctx, query, args...)
	}

	return dbx.DB.QueryContext(ctx, query, args...)
}

func (dbx *dbx) QueryRow(query string, args ...interface{}) *sql.Row {
	return dbx.DB.QueryRow(query, args...)
}

func (dbx *dbx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx)
	if ok {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return dbx.DB.QueryRowContext(ctx, query, args...)
}

func (dbx *dbx) Begin() (*sql.Tx, error) {
	return dbx.DB.Begin()
}

func (dbx *dbx) Close() error {
	return dbx.DB.Close()
}
