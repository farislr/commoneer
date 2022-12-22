package pkghttp

import (
	"context"
	"encoding/json"
	"net/http"
)

var (
	applicationJson = "application/json; charset=utf-8"
	contentType     = "Content-Type"
	fields          = "fields"
)

func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set(contentType, applicationJson)
	if headerer, ok := response.(Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	code := http.StatusOK
	if sc, ok := response.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}
	return json.NewEncoder(w).Encode(response)
}

func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	err = json.NewEncoder(w).Encode(err)
	if err != nil {
		return
	}
}

type StatusCoder interface {
	StatusCode() int
}

type Headerer interface {
	Headers() http.Header
}
