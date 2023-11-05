package nntpclient

import (
	"fmt"
	"net/textproto"
)

// Head retrieves only the headers for an article. The id parameter is
// handled in the same way as it is in [Article].
func (c *Client) Head(id string) (textproto.MIMEHeader, error) {
	var cmd string
	if id == "" {
		cmd = "HEAD"
	} else {
		cmd = fmt.Sprintf("HEAD %s", id)
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

	if code != 221 {
		return nil, UnexpectedError(code, message)
	}

	return c.readHeaders()
}
