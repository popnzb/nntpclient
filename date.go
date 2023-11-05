package nntpclient

import (
	"fmt"
	"time"
)

// Date retrieves the current date and time as it is known by the remote
// server.
func (c *Client) Date() (time.Time, error) {
	code, message, err := c.sendCommand("DATE")
	if err != nil {
		return time.Time{}, err
	}

	if code != 111 {
		return time.Time{}, UnexpectedError(code, message)
	}

	t, err := time.Parse("20060102150405", message)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date (%s): %v", message, err)
	}

	return t, nil
}
