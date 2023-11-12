package nntpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/textproto"
)

// ArticleAsBytes is a wrapper for [Article] that reads the whole article
// into a buffer before returning the headers and the body of the article as
// a slice of bytes.
func (c *Client) ArticleAsBytes(id string) (textproto.MIMEHeader, []byte, error) {
	var body bytes.Buffer
	headers, err := c.Article(id, &body)
	return headers, body.Bytes(), err
}

// Article gets an article from the server. The id parameter may be any of:
//
// 1. empty string (`""`) -- retrieve the "next" or "last" selected article
// 2. group article id -- the internal group article id, e.g. `1`
// 3. global id -- the global id for an article with brackets, e.g. `<foo.bar>`
//
// Note, when using the global id it is not necessary to select a group first.
// But when using an internal group id, or the empty string, a group must be
// selected prior to invoking this function.
//
// After reading the body of the article into the given writer, the headers
// will be returned if an error did not occur. If an error does occur, whatever
// part of the article body was read, if any, will have been written to the
// writer, nil will be returned for the headers, and the error will be returned.
func (c *Client) Article(id string, writer io.Writer) (textproto.MIMEHeader, error) {
	var cmd string
	if id == "" {
		cmd = "ARTICLE"
	} else {
		cmd = fmt.Sprintf("ARTICLE %s", id)
	}

	code, message, err := c.sendCommand(cmd)
	if err != nil {
		return nil, err
	}

	switch code {
	case 412:
		return nil, ErrNoGroupSelected
	case 420:
		return nil, ErrCurrentArticleNumInvalid
	case 423:
		return nil, ErrNoArticleWithNum
	case 430:
		return nil, ErrNoArticleWithId
	}

	if code != 220 {
		return nil, UnexpectedError(code, message)
	}

	headers, err := c.readHeaders()
	if err != nil {
		return nil, err
	}

	err = c.readBody(writer)
	if err != nil {
		return nil, err
	}

	return headers, nil
}
