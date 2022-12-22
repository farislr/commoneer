package internal

import (
	"testing"
	"time"
)

type modelDB struct {
	Name      string    `column:"name"`
	CreatedAt time.Time `column:"created_at"`
}

func TestModifyOrKeepField(t *testing.T) {
	var model modelDB
	var models []modelDB

	type args struct {
		existingQuery string
		model         interface{}
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
	}{
		{
			name: "1 model",
			args: args{
				existingQuery: "SELECT * FROM model",
				model:         model,
			},
			wantQuery: "SELECT name, created_at FROM model",
		},
		{
			name: "slice model",
			args: args{
				existingQuery: "SELECT * FROM model",
				model:         models,
			},
			wantQuery: "SELECT name, created_at FROM model",
		},
		{
			name: "no asterix",
			args: args{
				existingQuery: "SELECT name, created_at FROM model",
				model:         model,
			},
			wantQuery: "SELECT name, created_at FROM model",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuery := ModifyOrKeepField(tt.args.existingQuery, tt.args.model); gotQuery != tt.wantQuery {
				t.Errorf("ModifyOrKeepField() = %v, want %v", gotQuery, tt.wantQuery)
			}
		})
	}
}
