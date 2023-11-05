package nntpclient

import (
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Stat(t *testing.T) {
	badResponseHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		writeLines(c, "bad response")
	}

	errorCodesHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		switch strings.ToLower(params[0]) {
		case "1":
			writeLines(c, "412 no group selected")
		case "2":
			writeLines(c, "420 invalid article num")
		case "3":
			writeLines(c, "423 no article")
		case "4":
			writeLines(c, "430 no article")
		default:
			writeLines(c, "500 unrecognized command")
		}
	}

	emptyIdSuccessHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		assert.Equal(t, "stat", cmd)
		assert.Equal(t, 0, len(params))
		writeLines(c, "223 1 <foo.bar>")
	}

	withIdSuccessHandler := func(t *testing.T, c net.Conn, cmd string, params []string) {
		assert.Equal(t, "stat", cmd)
		assert.Equal(t, 1, len(params))
		assert.Equal(t, "<foo.bar>", params[0])
		writeLines(c, "223 1 <foo.bar>")
	}

	t.Run("handles bad response", func(t *testing.T) {
		server, client := getServerAndClient(t, badResponseHandler)
		defer server.Close()

		groupId, globalId, err := client.Stat("")
		assert.Equal(t, -1, groupId)
		assert.Equal(t, globalId, "")
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("handles possible error codes", func(t *testing.T) {
		server, client := getServerAndClient(t, errorCodesHandler)
		defer server.Close()

		groupId, globalId, err := client.Stat("1")
		assert.Equal(t, -1, groupId)
		assert.Equal(t, "", globalId)
		assert.Equal(t, true, errors.Is(err, ErrNoGroupSelected))

		groupId, globalId, err = client.Stat("2")
		assert.Equal(t, -1, groupId)
		assert.Equal(t, "", globalId)
		assert.Equal(t, true, errors.Is(err, ErrCurrentArticleNumInvalid))

		groupId, globalId, err = client.Stat("3")
		assert.Equal(t, -1, groupId)
		assert.Equal(t, "", globalId)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithNum))

		groupId, globalId, err = client.Stat("4")
		assert.Equal(t, -1, groupId)
		assert.Equal(t, "", globalId)
		assert.Equal(t, true, errors.Is(err, ErrNoArticleWithId))

		groupId, globalId, err = client.Stat("<foo.bar>")
		assert.Equal(t, -1, groupId)
		assert.Equal(t, "", globalId)
		assert.Equal(t, true, errors.Is(err, NntpError))
	})

	t.Run("handles success for no id", func(t *testing.T) {
		server, client := getServerAndClient(t, emptyIdSuccessHandler)
		defer server.Close()

		groupId, globalId, err := client.Stat("")
		assert.Nil(t, err)
		assert.Equal(t, 1, groupId)
		assert.Equal(t, "<foo.bar>", globalId)
	})

	t.Run("handles success with id", func(t *testing.T) {
		server, client := getServerAndClient(t, withIdSuccessHandler)
		defer server.Close()

		groupId, globalId, err := client.Stat("<foo.bar>")
		assert.Nil(t, err)
		assert.Equal(t, 1, groupId)
		assert.Equal(t, "<foo.bar>", globalId)
	})
}
