package nntpclient

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Next(t *testing.T) {
	badResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	error412Handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "412 error")
	}

	error420Handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "420 error")
	}

	error421Handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "421 error")
	}

	unexpectedErrorHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "404 missing")
	}

	successHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "223 next")
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, badResponseHandler)
		defer server.Close()

		err := client.Next()
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles 412 error", func(t *testing.T) {
		server, client := getServerAndClient(t, error412Handler)
		defer server.Close()

		err := client.Next()
		assert.Equal(t, true, errors.Is(err, ErrNoGroupSelected))
	})

	t.Run("handles 420 error", func(t *testing.T) {
		server, client := getServerAndClient(t, error420Handler)
		defer server.Close()

		err := client.Next()
		assert.Equal(t, true, errors.Is(err, ErrCurrentArticleNumInvalid))
	})

	t.Run("handles 421 error", func(t *testing.T) {
		server, client := getServerAndClient(t, error421Handler)
		defer server.Close()

		err := client.Next()
		assert.Equal(t, true, errors.Is(err, ErrNoNextArticle))
	})

	t.Run("handles unexpected error", func(t *testing.T) {
		server, client := getServerAndClient(t, unexpectedErrorHandler)
		defer server.Close()

		err := client.Next()
		assert.Equal(t, true, errors.Is(err, NntpError))
	})

	t.Run("handles success response", func(t *testing.T) {
		server, client := getServerAndClient(t, successHandler)
		defer server.Close()

		err := client.Next()
		assert.Nil(t, err)
	})
}
