package nntpclient

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

// Stat is used to determine if an article exists. It works like [Article]
// in that the id may be the empty string, an internal group id, or a global
// message id.
//
// If the article exists, its internal group id and global message id is
// returned. Otherwise, an error is returned along with `-1` and an empty
// string.
func (c *Client) Stat(id string) (int, string, error) {
	var cmd string
	if id == "" {
		cmd = "STAT"
	} else {
		cmd = fmt.Sprintf("STAT %s", id)
	}

	code, message, err := c.sendCommand(cmd)
	if err != nil {
		return -1, "", err
	}

	switch code {
	case 412:
		return -1, "", ErrNoGroupSelected
	case 420:
		return -1, "", ErrCurrentArticleNumInvalid
	case 423:
		return -1, "", ErrNoArticleWithNum
	case 430:
		return -1, "", ErrNoArticleWithId
	}

	if code != 223 {
		return -1, "", UnexpectedError(code, message)
	}

	parts := strings.Fields(message)

	return cast.ToInt(parts[0]), parts[1], nil
}
