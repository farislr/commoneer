package pkghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/suite"
)

type pkghttpSuite struct {
	router *httprouter.Router

	suite.Suite
}

func (ss *pkghttpSuite) SetupSuite() {
	r := httprouter.New()

	e1 := NewServer(func(ctx context.Context, req *Request[req1]) (resp *res1) {
		fmt.Printf("req.Body.Name: %v\n", req.Body.Name)

		return &res1{
			Code: 200,
			Msg:  "success " + req.Body.Name,
		}
	})

	r.Handler(http.MethodPost, "/endpoint/:id", e1)

	ss.router = r
}

func TestPkgHttpSuiteRun(t *testing.T) {
	suite.Run(t, new(pkghttpSuite))
}

type req1 struct {
	Name string `json:"name"`
}

type res1 struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (r *res1) Error() string {
	return r.Msg
}

func (ss *pkghttpSuite) TestEndpoint1() {
	type tests struct {
		desc string
		body req1
	}

	testCases := []tests{
		{
			desc: "success",
			body: req1{
				Name: "test",
			},
		},
	}
	for _, tt := range testCases {
		ss.Run(tt.desc, func() {
			b, err := json.Marshal(tt.body)
			ss.NoError(err)

			req := httptest.NewRequest("POST", "/endpoint/12", bytes.NewBuffer(b))
			w := httptest.NewRecorder()

			ss.router.ServeHTTP(w, req)

			fmt.Printf("w: %v\n", w)
		})
	}
}

func BenchmarkServeHTTP(b *testing.B) {
	e1 := NewServer(func(ctx context.Context, req *Request[req1]) (resp *res1) {
		// fmt.Printf("req.Body.Name: %v\n", req.Body.Name)

		return &res1{
			Code: 200,
			Msg:  "success " + req.Body.Name,
		}
	},
		WithPreRequestMiddleware(
			func(next EndpointFn[req1, *res1]) EndpointFn[req1, *res1] {
				return func(ctx context.Context, req *Request[req1]) (resp *res1) {
					return next(ctx, req)
				}
			},
		),
	)

	body := req1{
		Name: "test",
	}

	bb, err := json.Marshal(body)
	if err != nil {
		b.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/endpoint/12", bytes.NewBuffer(bb))
	w := httptest.NewRecorder()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e1.ServeHTTP(w, req)
	}
}

func BenchmarkNewServer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewServer(func(ctx context.Context, req *Request[req1]) (resp *res1) {
			// fmt.Printf("req.Body.Name: %v\n", req.Body.Name)

			return &res1{
				Code: 200,
				Msg:  "success " + req.Body.Name,
			}
		}, WithPreRequestMiddleware(
			func(next EndpointFn[req1, *res1]) EndpointFn[req1, *res1] {
				return func(ctx context.Context, req *Request[req1]) (resp *res1) {
					return next(ctx, req)
				}
			},
		))
	}
}
