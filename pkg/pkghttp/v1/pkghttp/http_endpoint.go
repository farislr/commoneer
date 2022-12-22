package pkghttp

import (
	"context"
	"net/http"
)

type responseSet interface {
	Error() string
}

type Request[T any] struct {
	Body   T
	Header http.Header

	req *http.Request
}

type EndpointFn[REQ any, RESP any] func(ctx context.Context, req *Request[REQ]) (resp RESP)

type EndpointMiddlewareFn[REQ any, RESP any] func(next EndpointFn[REQ, RESP]) EndpointFn[REQ, RESP]
