package pkghttp

import (
	"bytes"
	"io"
)

type reader struct {
	io.Reader
	readBuff *bytes.Buffer
	backBuff *bytes.Buffer
}

func newReader(r io.Reader) (*reader, error) {
	readBuff := bytes.Buffer{}
	if _, err := readBuff.ReadFrom(r); err != nil {
		return nil, err
	}
	backBuff := bytes.Buffer{}

	return &reader{
		Reader:   io.TeeReader(&readBuff, &backBuff),
		readBuff: &readBuff,
		backBuff: &backBuff,
	}, nil
}

func (r reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err == io.EOF {
		if err := r.reset(); err != nil {
			return 0, err
		}
	}

	return n, err
}

func (r reader) Close() error { return nil }

func (r reader) reset() error {
	_, err := io.Copy(r.readBuff, r.backBuff)

	return err
}
