package nntpclient

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// ListGroup represents a group information line from a list directive.
// See RFC 3977 §7.6.3.
type ListGroup struct {
	Name   string
	Low    int
	High   int
	Status string
}

// ListGroupTimes represents a group information line from a `active.times`
// command. See RFC 3977 §7.6.4.
type ListGroupTimes struct {
	Name    string
	Created time.Time
	Creator string
}

// ListDistribPattern represents distribution header values supported by the
// server. See RFC 3977 §7.6.5.
type ListDistribPattern struct {
	Weight  int
	Wildmat string
	Value   string
}

// ListNewsgroup represents a newsgroup information line from a
// `list newsgroups` directive. See RFC 3977 §7.6.6.
type ListNewsgroup struct {
	Name        string
	Description string
}

func (c *Client) listCmd(cmd string) ([]byte, error) {
	code, message, err := c.sendCommand(cmd)
	if err != nil {
		return nil, err
	}
	if code != 215 {
		return nil, UnexpectedError(code, message)
	}

	var body bytes.Buffer
	err = c.readBody(&body)

	return body.Bytes(), err
}

func bodyToListGroup(body []byte) map[string]ListGroup {
	result := make(map[string]ListGroup)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		result[parts[0]] = ListGroup{
			Name:   parts[0],
			Low:    cast.ToInt(parts[2]),
			High:   cast.ToInt(parts[1]),
			Status: parts[3],
		}
	}
	return result
}

// ListActive retrieves a list of active groups. The wildmat parameter can
// be the empty string to indicate "all groups."
func (c *Client) ListActive(wildmat string) (map[string]ListGroup, error) {
	var cmd string
	if wildmat == "" {
		cmd = "LIST ACTIVE"
	} else {
		cmd = fmt.Sprintf("LIST ACTIVE %s", wildmat)
	}

	body, err := c.listCmd(cmd)
	if err != nil {
		return nil, err
	}

	result := bodyToListGroup(body)

	return result, nil
}

// ListActiveTimes retrieves a list of groups, when they were created, and by
// whom. The wildmat parameter can be the empty string to indicate "all groups."
func (c *Client) ListActiveTimes(wildmat string) (map[string]ListGroupTimes, error) {
	var cmd string
	if wildmat == "" {
		cmd = "LIST ACTIVE.TIMES"
	} else {
		cmd = fmt.Sprintf("LIST ACTIVE.TIMES %s", wildmat)
	}

	body, err := c.listCmd(cmd)
	if err != nil {
		return nil, err
	}

	result := make(map[string]ListGroupTimes)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		result[parts[0]] = ListGroupTimes{
			Name:    parts[0],
			Created: time.Unix(cast.ToInt64(parts[1]), 0).UTC(),
			Creator: parts[2],
		}
	}

	return result, nil
}

// ListDistribPats retrieves a list of distribution header patterns supported
// by the server.
func (c *Client) ListDistribPats() ([]ListDistribPattern, error) {
	body, err := c.listCmd("LIST DISTRIB.PATS")
	if err != nil {
		return nil, err
	}

	result := make([]ListDistribPattern, 0)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		pattern := ListDistribPattern{
			Weight:  cast.ToInt(parts[0]),
			Wildmat: parts[1],
			Value:   parts[2],
		}
		result = append(result, pattern)
	}

	return result, nil
}

// ListNewsgroups retrieves a list of newsgroups known by the server. The
// wildmat parameter can be the empty string to indicate "all groups". The
// result is a map of group names to group names and group descriptions.
func (c *Client) ListNewsgroups(wildmat string) (map[string]ListNewsgroup, error) {
	var cmd string
	if wildmat == "" {
		cmd = "LIST NEWSGROUPS"
	} else {
		cmd = fmt.Sprintf("LIST NEWSGROUPS %s", wildmat)
	}

	body, err := c.listCmd(cmd)
	if err != nil {
		return nil, err
	}

	result := make(map[string]ListNewsgroup)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Bytes()
		sepIndex := bytes.IndexAny(line, " \t")
		name := string(line[0:sepIndex])
		value := strings.TrimSpace(string(line[sepIndex:]))
		result[name] = ListNewsgroup{Name: name, Description: value}
	}

	return result, nil
}
