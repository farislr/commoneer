package pkghttp

import (
	"encoding/json"
	"net/http"
)

type server[REQ any, RESP responseSet] struct {
	e EndpointFn[REQ, RESP]

	preRequestMiddlewares []EndpointMiddlewareFn[REQ, RESP]
}

func NewServer[REQ any, RESP responseSet](e EndpointFn[REQ, RESP], opts ...Option[REQ, RESP]) *server[REQ, RESP] {
	s := &server[REQ, RESP]{
		e: e,
	}

	for _, o := range opts {
		o.Apply(s)
	}

	return s
}

func (s *server[REQ, RESP]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body REQ
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return
	}

	req := &Request[REQ]{
		Body:   body,
		Header: r.Header,
		req:    r,
	}

	if s.preRequestMiddlewares != nil {
		for _, m := range s.preRequestMiddlewares {
			s.e = m(s.e)
		}
	}

	resp := s.e(ctx, req)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}

type Option[REQ any, RESP responseSet] interface {
	Apply(*server[REQ, RESP])
}

type optionFn[REQ any, RESP responseSet] func(*server[REQ, RESP])

func (fn optionFn[REQ, RESP]) Apply(s *server[REQ, RESP]) {
	fn(s)
}

func WithPreRequestMiddleware[REQ any, RESP responseSet](e ...EndpointMiddlewareFn[REQ, RESP]) Option[REQ, RESP] {
	return optionFn[REQ, RESP](func(s *server[REQ, RESP]) {
		s.preRequestMiddlewares = append(s.preRequestMiddlewares, e...)
	})
}
