package nntpclient

import (
	"time"
)

// NewGroups queries the server for the list of groups that have been added
// since a given time. If the location of that time is _not_ UTC, then only
// the date and time fields will be sent, leaving the interpretation of that
// time up to the server. If the location is set to UTC, then the GMT parameter
// is included in the query to the server.
func (c *Client) NewGroups(since time.Time) (map[string]ListGroup, error) {
	strTime := since.Format("20060102 150405")
	if since.Location() == time.UTC {
		strTime = strTime + " GMT"
	}

	code, message, err := c.sendCommand("NEWGROUPS " + strTime)
	if err != nil {
		return nil, err
	}
	if code != 231 {
		return nil, UnexpectedError(code, message)
	}

	body, err := c.readBody()
	if err != nil {
		return nil, err
	}

	result := bodyToListGroup(body)
	return result, nil
}
