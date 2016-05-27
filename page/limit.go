package page

import (
	"errors"
	"io"
)

// LimitedReader reads at most N bytes and returns an error when more than
// N bytes could have been read.
type LimitedReader struct {
	io.Reader
	N int64
}

// NewLimitedReadCloser wraps a Reader and limits the amount of bytes
// it can read before returning an error.
func NewLimitedReader(rc io.Reader, l int64) *LimitedReader {
	return &LimitedReader{rc, l}
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, errors.New("http: response body too large")
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.Reader.Read(p)
	l.N -= int64(n)
	return
}
