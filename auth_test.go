package nntpclient

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Authenticate(t *testing.T) {
	t.Run("handles user read error", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, cmd, "authinfo")
			assert.Equal(t, []string{"USER", "foo"}, params)
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		err := client.Authenticate("foo", "bar")
		assert.ErrorContains(t, err, "could not process response code")
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, cmd, "authinfo")
			assert.Equal(t, []string{"USER", "foo"}, params)
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		err := client.Authenticate("foo", "bar")
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response")
	})

	t.Run("handles pass read error", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			if params[0] == "USER" {
				writeLines(c, "381 user accepted")
				return
			}

			assert.Equal(t, "authinfo", cmd)
			assert.Equal(t, []string{"PASS", "bar"}, params)
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		err := client.Authenticate("foo", "bar")
		assert.ErrorContains(t, err, "could not process response code")
	})

	t.Run("handles pass unexpected error", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			if params[0] == "USER" {
				writeLines(c, "381 user accepted")
				return
			}

			assert.Equal(t, "authinfo", cmd)
			assert.Equal(t, []string{"PASS", "bar"}, params)
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		err := client.Authenticate("foo", "bar")
		assert.ErrorContains(t, err, "auth failed with code: 500 (boom)")
	})

	t.Run("handles success", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			if params[0] == "USER" {
				writeLines(c, "381 user accepted")
				return
			}

			assert.Equal(t, "authinfo", cmd)
			assert.Equal(t, []string{"PASS", "bar"}, params)
			writeLines(c, "281 pass accepted")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		err := client.Authenticate("foo", "bar")
		assert.Nil(t, err)
	})
}
