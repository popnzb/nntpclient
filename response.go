package nntpclient

import (
	"bufio"
	"io"
)

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
