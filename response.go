package nntpclient

import (
	"bufio"
	"io"
)

// MultibyteReader is an interface for readers that support reading
// multiple bytes up to a delimiter in one read. See [bufio.Reader].
type MultibyteReader interface {
	ReadBytes(delim byte) ([]byte, error)
}

type Response struct {
	bufferedReader *bufio.Reader
}

func NewResponse(conn io.ReadWriteCloser) *Response {
	return &Response{
		bufferedReader: bufio.NewReader(conn),
	}
}

func (r *Response) ReadBytes(delim byte) ([]byte, error) {
	return r.bufferedReader.ReadBytes(delim)
}
