package nntpclient

import (
	"errors"
	"net"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Head(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, headers)
		assert.ErrorContains(t, err, "could not process response code")
	})

	t.Run("handles 412 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "412 no group selected")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrNoGroupSelected))
	})

	t.Run("handles 420 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "420 invalid")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrCurrentArticleNumInvalid))
	})

	t.Run("handles 423 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "423 no num")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithNum))
	})

	t.Run("handles 430 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "430 no id")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithId))
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response code: 500 (boom)")
	})

	t.Run("handles success response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "221 head", "one: one", "two: two", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("foo")
		assert.Nil(t, err)

		expected := textproto.MIMEHeader{
			"One": {"one"},
			"Two": {"two"},
		}
		assert.Equal(t, expected, headers)
	})

	t.Run("handles success response (no id)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "221 head", "one: one", "two: two", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		headers, err := client.Head("")
		assert.Nil(t, err)

		expected := textproto.MIMEHeader{
			"One": {"one"},
			"Two": {"two"},
		}
		assert.Equal(t, expected, headers)
	})
}
