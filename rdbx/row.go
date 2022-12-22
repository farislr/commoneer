package rdbx

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

type Rows struct {
	*sql.Rows

	cachedRows *bytes.Buffer

	cache Cache

	queryKey string

	ctx context.Context
}

func (r *Rows) Close() error {
	if err := r.cache.Set(r.ctx, r.queryKey, r.cachedRows.Bytes()); err != nil {
		return err
	}

	return r.Rows.Close()
}

func (r *Rows) Next() bool {
	return r.Rows.Next()
}

func (r *Rows) Scan(dest ...interface{}) error {
	destDefault := make([]interface{}, len(dest))
	copy(destDefault, dest)

	if err := r.Rows.Scan(dest...); err != nil {
		return err
	}

	destVal, err := r.makeDestValue(destDefault)
	if err != nil {
		return err
	}

	d := r.joinBytes(destVal)
	fmt.Printf("d: %s\n", d)

	_, err = bytes.NewBuffer(d).WriteTo(r.cachedRows)
	if err != nil {
		return err
	}

	return nil
}

func (r *Rows) joinBytes(destVal [][]byte) []byte {
	d := bytes.Join(destVal, []byte(","))
	d = append(d, byte(';'))

	return d
}

func (r *Rows) makeDestValue(dest []interface{}) ([][]byte, error) {
	col, err := r.Rows.Columns()
	if err != nil {
		return nil, err
	}

	destVal := make([][]interface{}, len(col))
	for i := range col {
		reflectDestVal := reflect.ValueOf(destVal[i])
		reflectDestVal.Elem().Set(reflect.ValueOf(dest[i]))
	}

	return [][]byte{}, nil
}
