package dbmock

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/farislr/commoneer/rdbx/internal"
)

type DBTXMock interface {
	ExpectQueryx(expectedSQL string, model interface{}) *sqlmock.ExpectedQuery

	sqlmock.Sqlmock
}

type mock struct {
	sqlmock.Sqlmock
}

func New(sqlmock sqlmock.Sqlmock) *mock {
	return &mock{
		sqlmock,
	}
}

func (m *mock) ExpectQueryx(expectedSQL string, model interface{}) *sqlmock.ExpectedQuery {
	e := internal.ModifyOrKeepField(expectedSQL, model)
	return m.ExpectQuery(e)
}
