package socks5

import (
	"io"
)

type NoWritetoReader struct {
	readerwriter io.Reader
}

func NewNoWritetoReader(rw io.Reader) *NoWritetoReader {
	return &NoWritetoReader{readerwriter: rw}
}

func (l *NoWritetoReader) Read(p []byte) (n int, err error) {

	nr, err := (l.readerwriter).Read(p)
	return nr, err

}
