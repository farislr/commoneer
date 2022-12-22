package rdbx

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"reflect"

	"github.com/farislr/commoneer/rdbx/internal"
)

// contextKeyEnableSqlTx is a context key used to enable SQL transactions.
type contextKeyEnableSqlTx struct{}

// Queryx executes a query that returns rows, typically a SELECT, and maps the result to a struct or slice of structs.
// The model parameter should be a pointer to a struct or slice of structs.
// The query parameter can contain placeholders for arguments.
// The args parameter is a list of arguments to replace the placeholders in the query.
func (db *dbx) Queryx(
	ctx context.Context,
	query string,
	model interface{},
	args ...interface{},
) error {
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

	err = db.rowScan(rows, p, t)
	if err != nil {
		return err
	}

	return nil
}

// rowScan scans the rows returned by a query and maps the result to a struct or slice of structs.
func (db *dbx) rowScan(rows *Rows, rVal reflect.Value, rType reflect.Type) error {
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

		if err := rows.Scan(vPtrs...); err != nil {
			return err
		}

		if err := db.assignField(rType, el, columns, vs); err != nil {
			return err
		}

		if rVal.Kind() == reflect.Slice {
			rVal.Set(reflect.Append(rVal, el))
		}
	}

	return nil
}

// assignField assigns a value to a field in a struct based on the column name in the result set.
func (db *dbx) assignField(
	rType reflect.Type,
	rValue reflect.Value,
	columns []string,
	values []interface{},
) error {
	for i := 0; i < rType.NumField(); i++ {
		if c, ok := rType.Field(i).Tag.Lookup("column"); ok {
			for ii, col := range columns {
				if col == c {
					colVal := values[ii]

					if b, ok := colVal.([]byte); ok {
						colVal = string(b)
					}

					if ok := db.checkStructFieldSetable(rValue); ok {
						if err := db.checkStructFieldType(rValue.Field(i), colVal); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

// checkStructFieldType checks the type of a struct field and assigns a value to it.
func (db *dbx) checkStructFieldType(element reflect.Value, value interface{}) error {
	switch t := element.Addr().Interface().(type) {
	case sql.Scanner:
		if err := db.sqlScannerSet(element, t, value); err != nil {
			return err
		}
	default:
		element.Set(reflect.ValueOf(value).Convert(element.Type()))
	}

	return nil
}

// sqlScannerSet sets a value for a struct field that implements the sql.Scanner interface.
func (dbx *dbx) sqlScannerSet(element reflect.Value, t sql.Scanner, value interface{}) error {
	if err := t.Scan(value); err != nil {
		return err
	}

	element.Set(reflect.ValueOf(t).Elem())

	return nil
}

// checkStructFieldSetable checks if a struct field can be set.
func (dbx *dbx) checkStructFieldSetable(element reflect.Value) bool {
	if element.IsValid() && element.CanSet() && element.CanAddr() {
		return true
	}

	return false
}

// dbx is a wrapper around sql.DB that provides additional functionality.
type dbx struct {
	db *sql.DB

	cache Cache
}

// Begin starts a new transaction.
func (x *dbx) Begin() (*sql.Tx, error) {
	return x.db.Begin()
}

// Close closes the database connection.
func (x *dbx) Close() error {
	return x.db.Close()
}

// ExecContext executes a query that does not return rows, such as an INSERT or UPDATE.
// The query parameter can contain placeholders for arguments.
// The args parameter is a list of arguments to replace the placeholders in the query.
func (x *dbx) ExecContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (sql.Result, error) {
	if tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx); ok {
		return tx.ExecContext(ctx, query, args...)
	}

	return x.db.ExecContext(ctx, query, args...)
}

// PrepareContext prepares a statement for execution.
func (x *dbx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx); ok {
		return tx.PrepareContext(ctx, query)
	}

	return x.db.PrepareContext(ctx, query)
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The query parameter can contain placeholders for arguments.
// The args parameter is a list of arguments to replace the placeholders in the query.
// Returns a Rows object that wraps the result set.
// If the query has been executed before and the result set is cached, the cached result set will be returned.
func (x *dbx) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	queryKey := hex.EncodeToString([]byte(query))

	var rows *sql.Rows
	var err error

	res, err := x.cache.Get(ctx, string(queryKey))
	if err != nil {
		log.Printf("cache get error: %v", err)
	}

	if tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx); ok {
		rows, err = tx.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		goto ReturnRows
	}

	rows, err = x.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

ReturnRows:
	return &Rows{
		ctx:        ctx,
		Rows:       rows,
		cachedRows: bytes.NewBuffer(res),
		cache:      x.cache,
		queryKey:   queryKey,
	}, nil
}

// QueryRowContext executes a query that is expected to return at most one row.
// The query parameter can contain placeholders for arguments.
// The args parameter is a list of arguments to replace the placeholders in the query.
// Returns a Row object that wraps the result.
func (x *dbx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx, ok := ctx.Value(&contextKeyEnableSqlTx{}).(*sql.Tx); ok {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return x.db.QueryRowContext(ctx, query, args...)
}

// NewDbx creates a new dbx object.
func NewDbx(db *sql.DB, cache Cache) *dbx {
	return &dbx{
		db:    db,
		cache: cache,
	}
}
