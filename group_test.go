package nntpclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Group(t *testing.T) {
	t.Run("handles basic error", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &eofReader{}},
		}

		summary, err := c.Group("foo")
		assert.Nil(t, summary)
		assert.ErrorContains(t, err, "EOF")
	})

	t.Run("handles 411 response", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "411 no group\r\n"}},
		}

		summary, err := c.Group("foo")
		assert.Nil(t, summary)
		assert.Equal(t, true, errors.Is(err, ErrNoSuchGroup))
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "404 missing\r\n"}},
		}

		summary, err := c.Group("foo")
		assert.Nil(t, summary)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response code: 404 (missing)")
	})

	t.Run("return summary", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "211 1 2 3 foo\r\n"}},
		}

		summary, err := c.Group("foo")
		assert.Nil(t, err)

		expected := &GroupSummary{
			Name:   "foo",
			Number: 1,
			Low:    2,
			High:   3,
		}
		assert.Equal(t, expected, summary)
	})
}

func Test_ListGroup(t *testing.T) {
	t.Run("handles basic error", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &eofReader{}},
		}

		list, err := c.ListGroup("foo")
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "EOF")
	})

	t.Run("handles 411 response", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "411 no group\r\n"}},
		}

		list, err := c.ListGroup("foo")
		assert.Nil(t, list)
		assert.Equal(t, true, errors.Is(err, ErrNoSuchGroup))
	})

	t.Run("handles 412 response", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "412 no group selected\r\n"}},
		}

		list, err := c.ListGroup("foo")
		assert.Nil(t, list)
		assert.Equal(t, true, errors.Is(err, ErrNoGroupSelected))
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "404 missing\r\n"}},
		}

		list, err := c.ListGroup("foo")
		assert.Nil(t, list)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response code: 404 (missing)")
	})

	t.Run("returns error for bad body", func(t *testing.T) {
		c := Client{
			logger: NilLogger,
			conn: responseConn{response: &badBodyReader{
				initial: "211 1 2 3 foo",
			}},
		}

		list, err := c.ListGroup("foo")
		assert.Nil(t, list)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("returns a list", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &fullResponseReader{
				initial: "211 1 2 3 foo",
				payload: []string{"42", "43", "44", "."},
			}},
		}

		list, err := c.ListGroup("foo")
		assert.Nil(t, err)

		expected := &GroupList{
			GroupSummary: GroupSummary{
				Name:   "foo",
				Number: 1,
				Low:    2,
				High:   3,
			},
			ArticleNumbers: []int{42, 43, 44},
		}
		assert.Equal(t, expected, list)
	})
}
