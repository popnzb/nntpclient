package nntpclient

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/spf13/cast"
)

// GroupSummary represents the details about a group.
type GroupSummary struct {
	Name   string
	Number int
	Low    int
	High   int
}

// GroupList represents the details about a group along with the group local
// id numbers for the articles in the group.
type GroupList struct {
	GroupSummary
	ArticleNumbers []int
}

// Group selects a group and returns the summary for that group.
func (c *Client) Group(name string) (*GroupSummary, error) {
	code, message, err := c.sendCommand("GROUP " + name)
	if err != nil {
		return nil, err
	}
	if code == 411 {
		return nil, ErrNoSuchGroup
	}
	if code != 211 {
		return nil, UnexpectedError(code, message)
	}

	parts := strings.Fields(message)
	result := &GroupSummary{
		Name:   parts[3],
		Number: cast.ToInt(parts[0]),
		Low:    cast.ToInt(parts[1]),
		High:   cast.ToInt(parts[2]),
	}

	return result, nil
}

// ListGroup selects a group and returns a summary for the group along with
// a list of the group local article identifiers.
func (c *Client) ListGroup(name string) (*GroupList, error) {
	code, message, err := c.sendCommand("LISTGROUP " + name)
	if err != nil {
		return nil, err
	}
	if code == 411 {
		return nil, ErrNoSuchGroup
	}
	if code == 412 {
		return nil, ErrNoGroupSelected
	}
	if code != 211 {
		return nil, UnexpectedError(code, message)
	}

	parts := strings.Fields(message)
	result := &GroupList{
		GroupSummary: GroupSummary{
			Name:   parts[3],
			Number: cast.ToInt(parts[0]),
			Low:    cast.ToInt(parts[1]),
			High:   cast.ToInt(parts[2]),
		},
		ArticleNumbers: make([]int, 0),
	}

	var body bytes.Buffer
	err = c.readBody(&body)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(body.Bytes()))
	for scanner.Scan() {
		articleId := scanner.Text()
		result.ArticleNumbers = append(result.ArticleNumbers, cast.ToInt(articleId))
	}

	return result, nil
}
