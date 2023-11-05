package nntpclient

import (
	"fmt"
)

// Body is used to retrieve only the body of an article. The id parameter
// is handled in the same way as it is by [Article].
func (c *Client) Body(id string) ([]byte, error) {
	var cmd string
	if id == "" {
		cmd = "BODY"
	} else {
		cmd = fmt.Sprintf("BODY %s", id)
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

	if code != 222 {
		return nil, UnexpectedError(code, message)
	}

	return c.readBody()
}
