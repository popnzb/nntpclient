package nntpclient

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"time"
)

// NewNews queries the server for a list of new articles in groups matching
// the given wildmat since the given time. As with [NewGroups], the GMT
// parameter is dependent upon the since time being set to UTC.
func (c *Client) NewNews(wildmat string, since time.Time) ([]string, error) {
	if wildmat == "" {
		return nil, errors.New("wildmat cannot be empty")
	}

	strTime := since.Format("20060102 150405")
	if since.Location() == time.UTC {
		strTime = strTime + " GMT"
	}

	code, message, err := c.sendCommand(fmt.Sprintf("NEWNEWS %s %s", wildmat, strTime))
	if err != nil {
		return nil, err
	}
	if code != 230 {
		return nil, UnexpectedError(code, message)
	}

	body, err := c.readBody()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}

	return result, nil
}
