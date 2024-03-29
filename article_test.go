package nntpclient

import (
	"bytes"
	"errors"
	"net"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ArticleAsBytes(t *testing.T) {
	handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(
			c,
			"220 article",
			"header: one", "header: two",
			"",
			"an article",
			".",
		)
		c.Close()
	}

	server, client := getServerAndClient(t, handler)
	defer server.Close()

	headers, body, err := client.ArticleAsBytes("")
	assert.Nil(t, err)
	assert.Equal(t, "an article\r\n", string(body))

	expectedHeaders := textproto.MIMEHeader{"Header": {"one", "two"}}
	assert.Equal(t, expectedHeaders, headers)
}

func Test_Article(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.ErrorContains(t, err, "could not process response code:")
	})

	t.Run("handles 412 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "412 no group selected")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrNoGroupSelected))
	})

	t.Run("handles 420 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "420 invalid")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrCurrentArticleNumInvalid))
	})

	t.Run("handles 423 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "423 no num")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithNum))
	})

	t.Run("handles 430 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "430 no id")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithId))
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response code: 500 (boom)")
	})

	t.Run("handles error reading headers", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "220 article", "\tbad: header")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Empty(t, body)
		assert.Nil(t, headers)
		assert.ErrorContains(t, err, "malformed headers, found folded")
	})

	t.Run("handles error reading body", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(
				c,
				"220 article",
				"header: one", "header: two",
				"",
				"incomplete body",
			)
			c.Close()
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Nil(t, headers)
		assert.Equal(t, "incomplete body\r\n", body.String())
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("handles a success response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(
				c,
				"220 article",
				"header: one", "header: two",
				"",
				"an article",
				".",
			)
			c.Close()
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		headers, err := client.Article("foo", &body)
		assert.Nil(t, err)
		assert.Equal(t, "an article\r\n", body.String())

		expectedHeaders := textproto.MIMEHeader{"Header": {"one", "two"}}
		assert.Equal(t, expectedHeaders, headers)
	})
}
