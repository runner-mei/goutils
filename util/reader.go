package util

import "io"

var _ io.ReadCloser = ReadCloser{}

type ReadCloser struct {
	io.Reader
	io.Closer
}

func (rc ReadCloser) Close() error {
	if rc.Closer != nil {
		return rc.Closer.Close()
	}
	return nil
}

func ToReadCloser(r io.Reader) ReadCloser {
	rc := ReadCloser{Reader: r}
	if a, ok := r.(io.Closer); ok {
		rc.Closer = a
	}
	return rc
}
