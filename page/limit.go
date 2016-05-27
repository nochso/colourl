package page

import (
	"errors"
	"io"
)

// LimitedReadCloser reads at most N bytes and returns an error when more than
// N bytes could have been read.
type LimitedReadCloser struct {
	io.ReadCloser
	N int64
}

// NewLimitedReadCloser wraps a ReadCloser and limits the amount of bytes
// it can read before returning an error.
func NewLimitedReadCloser(rc io.ReadCloser, l int64) *LimitedReadCloser {
	return &LimitedReadCloser{rc, l}
}

func (l *LimitedReadCloser) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, errors.New("http: response body too large")
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.ReadCloser.Read(p)
	l.N -= int64(n)
	return
}
