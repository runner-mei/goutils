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

var _ io.WriteCloser = WriteCloser{}

type WriteCloser struct {
	io.Writer
	io.Closer
}

func (rc WriteCloser) Close() error {
	if rc.Closer != nil {
		return rc.Closer.Close()
	}
	return nil
}

func ToWriteCloser(w io.Writer) WriteCloser {
	wc := WriteCloser{Writer: w}
	if a, ok := w.(io.Closer); ok {
		wc.Closer = a
	}
	return wc
}

// ReadCloser returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
func TeeReadCloser(r io.ReadCloser, w io.Writer) io.ReadCloser {
	if r == nil {
		return r
	}
	return &teeReader{r, w}
}

type teeReader struct {
	r io.ReadCloser
	w io.Writer
}

func (t *teeReader) Close() error {
	return t.r.Close()
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}
