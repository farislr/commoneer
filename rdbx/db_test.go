package rdbx_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/farislr/commoneer/rdbx"
	"github.com/farislr/commoneer/rdbx/dbmock"
	"github.com/go-redis/redismock/v8"
	"github.com/go-redsync/redsync/v4"
	_rsyncpool "github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/stretchr/testify/suite"
)

type DBTXSuite struct {
	dbx rdbx.DBTX

	dbMock dbmock.DBTXMock
	rdMock redismock.ClientMock

	suite.Suite
}

func TestDBTXSuite(t *testing.T) {
	suite.Run(t, new(DBTXSuite))
}

func (s *DBTXSuite) SetupTest() {
	rawDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		s.Error(err)
	}

	s.dbMock = dbmock.New(mock)

	rd, rMock := redismock.NewClientMock()
	s.rdMock = rMock

	pool := _rsyncpool.NewPool(rd)
	rsync := redsync.New(pool)

	s.dbx = rdbx.NewDBx(rawDb, rd, rsync)
}

type transactionDate time.Time

func (t transactionDate) Assign(value interface{}) (out interface{}, err error) {
	v, ok := value.(time.Time)
	if !ok {
		return nil, errors.New("value is not time.Time")
	}

	return transactionDate(v), nil
}

type modelTable struct {
	ID        int64        `column:"id"`
	Name      string       `column:"name"`
	Date      time.Time    `column:"date"`
	CreatedAt sql.NullTime `column:"created_at"`
}

type customModelTable struct {
	ID        int64           `column:"id"`
	Name      string          `column:"name"`
	Date      transactionDate `column:"date"`
	CreatedAt sql.NullTime    `column:"created_at"`
	CreatedBy sql.NullString  `column:"created_by"`
	RefNumber sql.NullInt64   `column:"ref_number"`
}

func (s *DBTXSuite) Test_dbx_Queryx() {
	var model modelTable
	var models []modelTable

	var customModels []customModelTable

	type params struct {
		ctx   context.Context
		query string
		model interface{}
		args  []interface{}
	}

	tests := []struct {
		name     string
		params   params
		wantErr  bool
		mockRows *sqlmock.Rows
	}{
		{
			name: "SuccessQueryAll",
			params: params{
				ctx:   context.Background(),
				query: `SELECT * FROM model_table`,
				model: &model,
			},
			wantErr: false,
			mockRows: sqlmock.NewRows([]string{
				"id",
				"name",
				"date",
				"created_at",
			}).AddRow(1, "test name", time.Now(), time.Now()),
		},
		{
			name: "SuccessQueryArrayAll",
			params: params{
				ctx:   context.Background(),
				query: `SELECT * FROM model_table`,
				model: &models,
			},
			wantErr: false,
			mockRows: sqlmock.NewRows([]string{
				"id",
				"name",
				"date",
				"created_at",
			}).
				AddRow(1, "test name", time.Now(), time.Now()).
				AddRow(2, "test name 2", time.Now(), time.Now()),
		},
		{
			name: "SuccessQueryWithUnorderedField",
			params: params{
				ctx:   context.Background(),
				query: `SELECT name, date, id, created_at FROM model_table`,
				model: &model,
			},
			wantErr: false,
			mockRows: sqlmock.NewRows([]string{
				"name",
				"date",
				"id",
				"created_at",
			}).AddRow(
				"test name unordered",
				time.Now(),
				100,
				time.Now(),
			),
		},
		{
			name: "SuccessQueryWithCustomField",
			params: params{
				ctx:   context.Background(),
				query: `SELECT name, date, id, created_at FROM model_table`,
				model: &customModels,
			},
			wantErr: false,
			mockRows: sqlmock.NewRows([]string{
				"name",
				"date",
				"id",
				"created_at",
				"created_by",
				"ref_number",
			}).AddRow(
				"test name custom field",
				time.Now(),
				101,
				time.Now(),
				"people",
				123,
			).AddRow(
				"test name custom field 2",
				time.Now(),
				102,
				time.Now(),
				nil,
				nil,
			),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Logf("tt.params.model before: %v\n", tt.params.model)

			s.dbMock.ExpectQueryx(tt.params.query, tt.params.model).WillReturnRows(tt.mockRows)

			err := s.dbx.Queryx(tt.params.ctx, tt.params.query, tt.params.model, tt.params.args...)
			if (err != nil) != tt.wantErr {
				s.T().Log(err)

				s.Error(err)
			}

			s.NoError(err)

			err = s.dbMock.ExpectationsWereMet()
			if err != nil {
				s.T().Log(err)

				s.Error(err)
			}

			s.NoError(err)

			s.T().Logf("tt.params.model after: %v\n", tt.params.model)

			s.NotEmpty(tt.params.model)
		})
	}
}

func (s *DBTXSuite) Test_dbx_QueryTransactioner() {
	var now = time.Now()

	type params struct {
		ctx   context.Context
		query string
		model interface{}
		args  []interface{}
	}

	tests := []struct {
		name      string
		params    params
		wantErr   bool
		wantPanic bool
	}{
		{
			name: "Success",
			params: params{
				ctx:   context.Background(),
				query: `INSERT INTO model_table (id, name, created_at) VALUES (?, ?, ?)`,
				model: customModelTable{
					ID:        1,
					Name:      "test name",
					CreatedAt: sql.NullTime{Time: now, Valid: true},
				},
				args: []interface{}{
					int64(1),
					"test name",
					now,
				},
			},
			wantErr:   false,
			wantPanic: false,
		},
		{
			name: "Rollback",
			params: params{
				ctx:   context.Background(),
				query: `INSERT INTO model_table (id, name, created_at) VALUES (?, ?, ?)`,
				model: customModelTable{
					ID:        1,
					Name:      "test name",
					CreatedAt: sql.NullTime{Time: now, Valid: true},
				},
				args: []interface{}{
					int64(1),
					"test name",
					now,
				},
			},
			wantErr: true,
		},
		{
			name: "PanicRollback",
			params: params{
				ctx:   context.Background(),
				query: `INSERT INTO model_table (id, name, created_at) VALUES (?, ?, ?)`,
				model: customModelTable{
					ID:        1,
					Name:      "test name",
					CreatedAt: sql.NullTime{Time: now, Valid: true},
				},
				args: []interface{}{
					int64(1),
					"test name",
					now,
				},
			},
			wantErr:   true,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := rdbx.NewTransactioner(s.dbx)

			switch m := tt.params.model.(type) {
			case customModelTable:
				s.rdMock.ExpectSet("idempotent-key", m.ID, 5*time.Minute).SetVal(fmt.Sprint(m.ID))
				defer s.rdMock.ClearExpect()

				if tt.wantErr {
					s.dbMock.ExpectBegin()
					s.dbMock.ExpectExec(tt.params.query).WithArgs(m.ID, m.Name, m.CreatedAt).WillReturnError(sql.ErrNoRows)
					s.dbMock.ExpectRollback()
				} else {
					s.dbMock.ExpectBegin()
					s.dbMock.ExpectExec(tt.params.query).WithArgs(m.ID, m.Name, m.CreatedAt).WillReturnResult(sqlmock.NewResult(1, 1))
					s.dbMock.ExpectCommit()
				}

			}

			err := d.EnableTx(tt.params.ctx, func(ctx context.Context) error {
				err := s.dbx.Set(ctx, "idempotent-key", int64(1), 5*time.Minute).Err()
				if err != nil {
					s.T().Log(err)

					s.Error(err)
					return err
				}
				s.NoError(err)

				res, err := s.dbx.ExecContext(ctx, tt.params.query, tt.params.args...)
				if (err != nil) || tt.wantErr {
					s.T().Log(err)

					s.Error(err)

					if tt.wantPanic {
						panic(err)
					}

					return err
				}
				s.NoError(err)

				s.T().Log(res)

				return nil
			})
			if (err != nil) || tt.wantErr {
				if tt.wantPanic {
					s.Panics(func() {
						panic("there is a panic")
					})
					return
				}

				s.T().Log(err)

				s.Error(err)
				return
			}
			s.NoError(err)

			err = s.dbMock.ExpectationsWereMet()
			s.NoError(err)
		})
	}
}
