package nntpclient

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Quit(t *testing.T) {
	badResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	successHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "205 closing")
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, badResponseHandler)
		defer server.Close()

		err := client.Quit()
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles success", func(t *testing.T) {
		server, client := getServerAndClient(t, successHandler)
		defer server.Close()

		err := client.Quit()
		assert.Nil(t, err)
	})

	t.Run("close alias succeeds", func(t *testing.T) {
		server, client := getServerAndClient(t, successHandler)
		defer server.Close()

		err := client.Close()
		assert.Nil(t, err)
	})
}
