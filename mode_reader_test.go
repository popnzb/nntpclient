package nntpclient

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeReader(t *testing.T) {
	badResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	unexpectedCodeHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "404 missing")
	}

	err502Handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "502 nope")
	}

	success200Handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "200 reader")
	}

	success201Handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "201 reader")
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, badResponseHandler)
		defer server.Close()

		err := client.ModeReader()
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles unexpected response code", func(t *testing.T) {
		server, client := getServerAndClient(t, unexpectedCodeHandler)
		defer server.Close()

		err := client.ModeReader()
		assert.ErrorContains(t, err, "unexpected response code")
	})

	t.Run("handles 502 response", func(t *testing.T) {
		server, client := getServerAndClient(t, err502Handler)
		defer server.Close()

		err := client.ModeReader()
		assert.Equal(t, true, errors.Is(err, ErrReadingUnavailable))
	})

	t.Run("handles 200 response", func(t *testing.T) {
		server, client := getServerAndClient(t, success200Handler)
		defer server.Close()

		err := client.ModeReader()
		assert.Nil(t, err)
		assert.Equal(t, true, client.CanPost)
	})

	t.Run("handles 201 response", func(t *testing.T) {
		server, client := getServerAndClient(t, success201Handler)
		defer server.Close()

		err := client.ModeReader()
		assert.Nil(t, err)
		assert.Equal(t, false, client.CanPost)
	})
}
