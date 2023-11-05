package nntpclient

import (
	"fmt"
	"net/textproto"
)

// Article gets an article from the server. The id parameter may be any of:
//
// 1. empty string (`""`) -- retrieve the "next" or "last" selected article
// 2. group article id -- the internal group article id, e.g. `1`
// 3. global id -- the global id for an article with brackets, e.g. `<foo.bar>`
//
// Note, when using the global id it is not necessary to select a group first.
// But when using a internal group id, or the empty string, a group must be
// selected prior to invoking this function.
//
// The result is a set of article headers and the bytes representing the article
// body. If an error occurs at any point while processing the article, only
// the error will be returned.
func (c *Client) Article(id string) (textproto.MIMEHeader, []byte, error) {
	var cmd string
	if id == "" {
		cmd = "ARTICLE"
	} else {
		cmd = fmt.Sprintf("ARTICLE %s", id)
	}

	code, message, err := c.sendCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	switch code {
	case 412:
		return nil, nil, ErrNoGroupSelected
	case 420:
		return nil, nil, ErrCurrentArticleNumInvalid
	case 423:
		return nil, nil, ErrNoArticleWithNum
	case 430:
		return nil, nil, ErrNoArticleWithId
	}

	if code != 220 {
		return nil, nil, UnexpectedError(code, message)
	}
	c.logger.Debug("article", "message", message)

	headers, err := c.readHeaders()
	if err != nil {
		return nil, nil, err
	}

	body, err := c.readBody()
	if err != nil {
		return nil, nil, err
	}

	return headers, body, nil
}
