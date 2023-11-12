package nntpclient

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Capabilities(t *testing.T) {
	badResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	unexpectedErrorHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "404 missing")
	}

	bodyReadErrHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "101 capabilities", "broken")
		c.Close()
	}

	fullBodyHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		capsBody := []string{
			"101 capabilities",
			"VERSION 2",
			"READER",
			"LIST ACTIVE NEWSGROUPS",
			".",
		}
		writeLines(c, capsBody...)
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, badResponseHandler)
		defer server.Close()

		caps, err := client.Capabilities()
		assert.Nil(t, caps)
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		server, client := getServerAndClient(t, unexpectedErrorHandler)
		defer server.Close()

		caps, err := client.Capabilities()
		assert.Nil(t, caps)
		assert.Equal(t, true, errors.Is(err, NntpError))
	})

	t.Run("handles a broken body read", func(t *testing.T) {
		server, client := getServerAndClient(t, bodyReadErrHandler)
		defer server.Close()

		caps, err := client.Capabilities()
		assert.Nil(t, caps)
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("reads a standard body", func(t *testing.T) {
		server, client := getServerAndClient(t, fullBodyHandler)
		defer server.Close()

		caps, err := client.Capabilities()
		assert.Nil(t, err)

		expected := &Capabilities{
			"VERSION": {"2"},
			"READER":  {},
			"LIST":    {"ACTIVE", "NEWSGROUPS"},
		}
		assert.Equal(t, expected, caps)
	})
}
