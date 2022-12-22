package pkghttp

import (
	"context"
	"net/http"
)

type ResponseEncoderFn func(ctx context.Context, w http.ResponseWriter, response interface{}) error

type ErrorResponseEncoderFn func(ctx context.Context, err error, w http.ResponseWriter)
