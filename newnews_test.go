package nntpclient

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewNews(t *testing.T) {
	// 2023-11-12T08:00:00.000-05:00
	baseDate := time.Unix(1699794000, 0)

	t.Run("returns error for empty wildmat", func(t *testing.T) {
		client := Client{}
		list, err := client.NewNews("", time.Now())
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "wildmat cannot be empty")
	})

	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.NewNews("*", baseDate)
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "could not process")
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.NewNews("*", baseDate)
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "unexpected response code: 500 (boom)")
	})

	t.Run("handles bad body read", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "230 list", "bad body")
			c.Close()
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.NewNews("*", baseDate)
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("returns list (ambiguous time)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "newnews", cmd)
			assert.Equal(t, []string{"*", "20231112", "080000"}, params)
			writeLines(c, "230 list", "<article1>", "<article2>", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.NewNews("*", baseDate)
		assert.Nil(t, err)

		expected := []string{"<article1>", "<article2>"}
		assert.Equal(t, expected, list)
	})

	t.Run("returns list (utc time)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "newnews", cmd)
			assert.Equal(t, []string{"*", "20231112", "130000", "GMT"}, params)
			writeLines(c, "230 list", "<article1>", "<article2>", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.NewNews("*", baseDate.UTC())
		assert.Nil(t, err)

		expected := []string{"<article1>", "<article2>"}
		assert.Equal(t, expected, list)
	})
}
