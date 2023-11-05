package nntpclient

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Date(t *testing.T) {
	invalidResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	unexpectedErrorHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "404 missing")
	}

	badDateHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "111 bad-date")
	}

	successHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		now, err := time.Parse(time.RFC3339, "2023-11-05T08:00:00-05:00")
		if err != nil {
			fmt.Println(err)
		}
		formatted := now.UTC().Format("20060102150405")
		writeLines(c, "111 "+formatted)
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, invalidResponseHandler)
		defer server.Close()

		date, err := client.Date()
		assert.Equal(t, time.Time{}, date)
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles unrecognized code", func(t *testing.T) {
		server, client := getServerAndClient(t, unexpectedErrorHandler)
		defer server.Close()

		date, err := client.Date()
		assert.Equal(t, time.Time{}, date)
		assert.ErrorContains(t, err, "unexpected")
	})

	t.Run("handles invalid date response", func(t *testing.T) {
		server, client := getServerAndClient(t, badDateHandler)
		defer server.Close()

		date, err := client.Date()
		assert.Equal(t, time.Time{}, date)
		assert.ErrorContains(t, err, "could not parse date")
	})

	t.Run("handles success response", func(t *testing.T) {
		server, client := getServerAndClient(t, successHandler)
		defer server.Close()

		date, err := client.Date()
		require.Nil(t, err)
		assert.Equal(t, "2023-11-05 13:00:00 +0000 UTC", date.String())
	})
}
