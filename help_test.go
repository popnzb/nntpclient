package nntpclient

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Help(t *testing.T) {
	badResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	unexpectedCodeHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "404 missing")
	}

	badBodyReadHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "100 broken")
		c.Close()
	}

	successHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "100 help", "this is", "some help", ".")
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, badResponseHandler)
		defer server.Close()

		help, err := client.Help()
		assert.Equal(t, "", help)
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles unexpected response code", func(t *testing.T) {
		server, client := getServerAndClient(t, unexpectedCodeHandler)
		defer server.Close()

		help, err := client.Help()
		assert.Equal(t, "", help)
		assert.ErrorContains(t, err, "unexpected response code")
	})

	t.Run("handles bad body read", func(t *testing.T) {
		server, client := getServerAndClient(t, badBodyReadHandler)
		defer server.Close()

		help, err := client.Help()
		assert.Equal(t, "", help)
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("handles a good response", func(t *testing.T) {
		server, client := getServerAndClient(t, successHandler)
		defer server.Close()

		expected := "this is\r\nsome help\r\n"
		help, err := client.Help()
		assert.Nil(t, err)
		assert.Equal(t, expected, help)
	})
}
