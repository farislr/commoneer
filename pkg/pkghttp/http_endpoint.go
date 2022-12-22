package pkghttp

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Request interface {
	AddCookie(c *http.Cookie)
	BasicAuth() (username string, password string, ok bool)
	Clone(ctx context.Context) *http.Request
	Context() context.Context
	Cookie(name string) (*http.Cookie, error)
	Cookies() []*http.Cookie
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	FormValue(key string) string
	MultipartReader() (*multipart.Reader, error)
	ParseForm() error
	ParseMultipartForm(maxMemory int64) error
	PostFormValue(key string) string
	ProtoAtLeast(major int, minor int) bool
	Referer() string
	SetBasicAuth(username string, password string)
	UserAgent() string
	WithContext(ctx context.Context) *http.Request
	Write(w io.Writer) error
	WriteProxy(w io.Writer) error

	Body() io.ReadCloser
	Decode(v any) error
	Header() http.Header
	URL() *url.URL
}

type request struct {
	*http.Request
}

func (r *request) Body() io.ReadCloser {
	var err error
	if r.Request.Body != nil {
		r.Request.Body, err = newReader(r.Request.Body)
		if err != nil {
			return nil
		}
	}

	return r.Request.Body
}

func (r *request) Decode(v interface{}) error {
	b, err := io.ReadAll(r.Body())
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (r *request) Header() http.Header {
	return r.Request.Header
}

func (r *request) URL() *url.URL {
	return r.Request.URL
}

type EndpointMiddlewareFn func(next EndpointFn) EndpointFn

type EndpointFn func(ctx context.Context, r Request) (response interface{}, err error)

func (s *server) Server(e EndpointFn, options ...EndpointOption) *endpoint {
	ee := &endpoint{
		e:                    e,
		responseEncoder:      s.responseEncoder,
		errorResponseEncoder: s.errorResponseEncoder,
	}

	for _, o := range options {
		o.Apply(ee)
	}

	return ee
}

type EndpointOption interface {
	Apply(*endpoint)
}

type EndpointOptionFn func(*endpoint)

func (o EndpointOptionFn) Apply(e *endpoint) {
	o(e)
}

func WithContext(ContextFn ContextFn) EndpointOption {
	return EndpointOptionFn(func(e *endpoint) {
		e.ctxFn = ContextFn
	})
}

func WithPreRequestMiddleware(mds ...EndpointMiddlewareFn) EndpointOption {
	return EndpointOptionFn(func(e *endpoint) {
		e.preRequestMiddlewares = append(e.preRequestMiddlewares, mds...)
	})
}

type endpoint struct {
	e EndpointFn

	responseEncoder      ResponseEncoderFn
	errorResponseEncoder ErrorResponseEncoderFn

	ctxFn                 ContextFn
	preRequestMiddlewares []EndpointMiddlewareFn
}

func (s *endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := &request{r}

	if s.ctxFn != nil {
		ctx = s.ctxFn(ctx, req)
	}

	if s.preRequestMiddlewares != nil {
		for _, m := range s.preRequestMiddlewares {
			s.e = m(s.e)
		}
	}

	res, err := s.e(ctx, req)
	if err != nil {
		s.errorResponseEncoder(ctx, err, w)
		return
	}

	if err := s.responseEncoder(ctx, w, res); err != nil {
		s.errorResponseEncoder(ctx, err, w)
		return
	}
}
