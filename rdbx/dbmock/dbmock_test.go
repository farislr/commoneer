package dbmock

import (
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type modelDB struct {
	Name      string    `column:"name"`
	CreatedAt time.Time `column:"created_at"`
}

func Test_mock_ExpectQueryx(t *testing.T) {
	var model modelDB

	_, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}

	dbmock := New(mock)

	type fields struct {
		Sqlmock DBTXMock
	}
	type args struct {
		expectedSQL string
		model       interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(string, args, fields) *sqlmock.ExpectedQuery
	}{
		{
			name: "success",
			fields: fields{
				Sqlmock: dbmock,
			},
			args: args{
				expectedSQL: "SELECT * FROM model",
				model:       model,
			},
			want: func(q string, a args, f fields) *sqlmock.ExpectedQuery {
				return f.Sqlmock.ExpectQueryx(q, a.model)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.fields.Sqlmock)

			m.GetColumns(tt.args.model)

			if got := m.ExpectQueryx(tt.args.expectedSQL, tt.args.model); !reflect.DeepEqual(
				got,
				tt.want(tt.args.expectedSQL, tt.args, tt.fields),
			) {
				t.Errorf(
					"mock.ExpectQueryx() = %v, want %v",
					got,
					tt.want(tt.args.expectedSQL, tt.args, tt.fields),
				)
			}
		})
	}
}
