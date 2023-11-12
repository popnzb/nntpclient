package nntpclient

import (
	"bytes"
	"fmt"
	"io"
)

// BodyAsBytes is used to retrieve only the body of an article as a slice of
// bytes. The full body is read into the slice before it is returned. The id
// parameter is handled in the same way as it is by [Article].
func (c *Client) BodyAsBytes(id string) ([]byte, error) {
	var body bytes.Buffer
	err := c.Body(id, &body)
	return body.Bytes(), err
}

// Body is used to retrieve only the body of an article. As the article is
// read from the connection it is written to writer. The id parameter is
// handled in the same way as it is by [Article].
func (c *Client) Body(id string, writer io.Writer) error {
	var cmd string
	if id == "" {
		cmd = "BODY"
	} else {
		cmd = fmt.Sprintf("BODY %s", id)
	}

	code, message, err := c.sendCommand(cmd)
	if err != nil {
		return err
	}

	switch code {
	case 412:
		return ErrNoGroupSelected
	case 420:
		return ErrCurrentArticleNumInvalid
	case 423:
		return ErrNoArticleWithNum
	case 430:
		return ErrNoArticleWithId
	}

	if code != 222 {
		return UnexpectedError(code, message)
	}

	err = c.readBody(writer)
	return err
}
