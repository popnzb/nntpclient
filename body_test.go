package nntpclient

import (
	"bytes"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BodyAsBytes(t *testing.T) {
	handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "222 body", "one", "two", ".")
	}

	server, client := getServerAndClient(t, handler)
	defer server.Close()

	body, err := client.BodyAsBytes("")
	assert.Nil(t, err)
	assert.Equal(t, "one\r\ntwo\r\n", string(body))
}

func Test_Body(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Empty(t, body)
		assert.ErrorContains(t, err, "could not process response code")
	})

	t.Run("handles 412 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "412 no group selected")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Empty(t, body)
		assert.Equal(t, true, errors.Is(err, ErrNoGroupSelected))
	})

	t.Run("handles 420 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "420 invalid")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Empty(t, body)
		assert.Equal(t, true, errors.Is(err, ErrCurrentArticleNumInvalid))
	})

	t.Run("handles 423 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "423 no num")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Empty(t, body)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithNum))
	})

	t.Run("handles 430 response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "430 no id")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Empty(t, body)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithId))
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Empty(t, body)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response code: 500 (boom)")
	})

	t.Run("handles success response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "222 body", "one", "two", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("foo", &body)
		assert.Nil(t, err)
		assert.Equal(t, "one\r\ntwo\r\n", body.String())
	})

	t.Run("handles success response (no id)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "222 body", "one", "two", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		var body bytes.Buffer
		err := client.Body("", &body)
		assert.Nil(t, err)
		assert.Equal(t, "one\r\ntwo\r\n", body.String())
	})
}
