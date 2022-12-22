package dbmock

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/farislr/commoneer/rdbx/internal"
)

type DBTXMock interface {
	ExpectQueryx(expectedSQL string, model interface{}) *sqlmock.ExpectedQuery
	GetColumns(model interface{}) []string

	sqlmock.Sqlmock
}

type mock struct {
	sqlmock.Sqlmock
}

func New(sqlmock sqlmock.Sqlmock) DBTXMock {
	return &mock{
		sqlmock,
	}
}

func (m *mock) ExpectQueryx(expectedSQL string, model interface{}) *sqlmock.ExpectedQuery {
	e := internal.ModifyOrKeepField(expectedSQL, model)
	return m.ExpectQuery(e)
}

func (m mock) GetColumns(model interface{}) []string {
	return internal.GetColumns(model)
}
