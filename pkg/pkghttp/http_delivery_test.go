package pkghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_server_Server(t *testing.T) {
	type fields struct {
		responseEncoder      ResponseEncoderFn
		errorResponseEncoder ErrorResponseEncoderFn
	}
	type args struct {
		e EndpointFn
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "success",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
					return nil
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
			},
			args: args{
				e: func(ctx context.Context, r Request) (interface{}, error) {
					return nil, nil
				},
			},
		},
		{
			name: "endpoint error",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
					return nil
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
			},
			args: args{
				e: func(ctx context.Context, r Request) (interface{}, error) {
					return nil, errors.New("test endpoint error")
				},
			},
		},
		{
			name: "encode response error",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
					return errors.New("error while encode response")
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
			},
			args: args{
				e: func(ctx context.Context, r Request) (interface{}, error) {
					return nil, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(
				WithResponseEncoder(tt.fields.responseEncoder),
				WithErrorResponseEncoder(tt.fields.errorResponseEncoder),
			)

			checkFeatureFlag := func(next EndpointFn) EndpointFn {
				return func(ctx context.Context, r Request) (interface{}, error) {
					// do something before request on endpoint

					return next(ctx, r)
				}
			}

			passContext := func(ctx context.Context, r Request) context.Context {
				// do something to propagate context

				return ctx
			}

			opts := []EndpointOption{
				WithPreRequestMiddleware(
					checkFeatureFlag,
				),
				WithContext(
					passContext,
				),
			}

			e := s.Server(tt.args.e, opts...)

			type exReqStruct struct {
				Msg string `json:"msg"`
			}

			exReq := exReqStruct{
				Msg: "test",
			}

			b, err := json.Marshal(exReq)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/example", bytes.NewBuffer(b))
			if err != nil {
				assert.Error(t, err)
			}

			r := &request{req}

			var exReq2 exReqStruct
			err = r.Decode(&exReq2)
			fmt.Printf("err: %v\n", err)
			fmt.Printf("exReq2: %+v\n", exReq2)

			var exReq3 exReqStruct
			err = r.Decode(&exReq3)
			fmt.Printf("err: %v\n", err)
			fmt.Printf("exReq3: %v\n", exReq3)

			rr := httptest.NewRecorder()
			e.ServeHTTP(rr, req)
		})
	}
}
