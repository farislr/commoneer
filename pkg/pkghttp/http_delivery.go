package pkghttp

type server struct {
	responseEncoder      ResponseEncoderFn
	errorResponseEncoder ErrorResponseEncoderFn
}

func NewServer(options ...ServerOption) *server {
	s := &server{}

	defaults(s)

	for _, o := range options {
		o.Apply(s)
	}

	return s
}

func defaults(s *server) {
	s.responseEncoder = EncodeResponse
	s.errorResponseEncoder = ErrorEncoder
}

type ServerOption interface {
	Apply(*server)
}

type ServerOptionFn func(*server)

func (o ServerOptionFn) Apply(s *server) {
	o(s)
}

func WithResponseEncoder(e ResponseEncoderFn) ServerOption {
	return ServerOptionFn(func(s *server) {
		s.responseEncoder = e
	})
}

func WithErrorResponseEncoder(e ErrorResponseEncoderFn) ServerOption {
	return ServerOptionFn(func(s *server) {
		s.errorResponseEncoder = e
	})
}
