package nntpclient

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_listCmd(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		body, err := client.listCmd("list")
		assert.Nil(t, body)
		assert.ErrorContains(t, err, "could not process")
	})

	t.Run("handles unexpected response code", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "500 boom")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		body, err := client.listCmd("list")
		assert.Nil(t, body)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected response code: 500 (boom)")
	})

	t.Run("returns body bytes", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "215 list", "body line", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		body, err := client.listCmd("list")
		assert.Nil(t, err)
		assert.Equal(t, "body line\r\n", string(body))
	})
}

func Test_bodyToListGroup(t *testing.T) {
	input := []byte("a.group 42 1 y\r\nb.group 2 1 n\r\n")
	expected := map[string]ListGroup{
		"a.group": {Name: "a.group", Low: 1, High: 42, Status: "y"},
		"b.group": {Name: "b.group", Low: 1, High: 2, Status: "n"},
	}
	found := bodyToListGroup(input)
	assert.Equal(t, expected, found)
}

func Test_ListActive(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListActive("")
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "could not process")
	})

	t.Run("processes all groups request", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "215 list", "a.group 42 1 y", "b.group 2 1 n", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListActive("")
		assert.Nil(t, err)

		expected := map[string]ListGroup{
			"a.group": {Name: "a.group", Low: 1, High: 42, Status: "y"},
			"b.group": {Name: "b.group", Low: 1, High: 2, Status: "n"},
		}
		assert.Equal(t, expected, list)
	})

	t.Run("processes wildmat request", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "list", cmd)
			assert.Equal(t, []string{"ACTIVE", "*.group"}, params)
			writeLines(c, "215 list", "a.group 42 1 y", "b.group 2 1 n", ".")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListActive("*.group")
		assert.Nil(t, err)

		expected := map[string]ListGroup{
			"a.group": {Name: "a.group", Low: 1, High: 42, Status: "y"},
			"b.group": {Name: "b.group", Low: 1, High: 2, Status: "n"},
		}
		assert.Equal(t, expected, list)
	})
}

func Test_ListActiveTimes(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListActiveTimes("")
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "could not process")
	})

	t.Run("returns a list (no wildmat)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "list", cmd)
			assert.Equal(t, []string{"ACTIVE.TIMES"}, params)
			writeLines(
				c,
				"215 list",
				"misc.test 930445408 <creatme@isc.org>",
				"alt.rfc-writers.recovery 930562309 <m@example.com>",
				".",
			)
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListActiveTimes("")
		assert.Nil(t, err)

		expected := map[string]ListGroupTimes{
			"misc.test": {
				Name:    "misc.test",
				Created: time.Unix(930445408, 0).UTC(),
				Creator: "<creatme@isc.org>",
			},
			"alt.rfc-writers.recovery": {
				Name:    "alt.rfc-writers.recovery",
				Created: time.Unix(930562309, 0).UTC(),
				Creator: "<m@example.com>",
			},
		}
		assert.Equal(t, expected, list)
	})

	t.Run("returns a list (with wildmat)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "list", cmd)
			assert.Equal(t, []string{"ACTIVE.TIMES", "*"}, params)
			writeLines(
				c,
				"215 list",
				"misc.test 930445408 <creatme@isc.org>",
				"alt.rfc-writers.recovery 930562309 <m@example.com>",
				".",
			)
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListActiveTimes("*")
		assert.Nil(t, err)

		expected := map[string]ListGroupTimes{
			"misc.test": {
				Name:    "misc.test",
				Created: time.Unix(930445408, 0).UTC(),
				Creator: "<creatme@isc.org>",
			},
			"alt.rfc-writers.recovery": {
				Name:    "alt.rfc-writers.recovery",
				Created: time.Unix(930562309, 0).UTC(),
				Creator: "<m@example.com>",
			},
		}
		assert.Equal(t, expected, list)
	})
}

func Test_ListDistribPats(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListDistribPats()
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "could not process")
	})

	t.Run("returns a list", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "list", cmd)
			assert.Equal(t, []string{"DISTRIB.PATS"}, params)
			writeLines(
				c,
				"215 list",
				"10:local.*:local",
				"5:*:world",
				".",
			)
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListDistribPats()
		assert.Nil(t, err)

		expected := []ListDistribPattern{
			{
				Weight:  10,
				Wildmat: "local.*",
				Value:   "local",
			},
			{
				Weight:  5,
				Wildmat: "*",
				Value:   "world",
			},
		}
		assert.Equal(t, expected, list)
	})
}

func Test_ListNewsgroups(t *testing.T) {
	t.Run("handles bad response", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			writeLines(c, "bad response")
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListNewsgroups("")
		assert.Nil(t, list)
		assert.ErrorContains(t, err, "could not process")
	})

	t.Run("returns a list (no wildmat)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "list", cmd)
			assert.Equal(t, []string{"NEWSGROUPS"}, params)
			writeLines(
				c,
				"215 list",
				"misc.test General Usenet testing",
				"alt.rfc-writers.recovery RFC Writers Recovery",
				".",
			)
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListNewsgroups("")
		assert.Nil(t, err)

		expected := map[string]ListNewsgroup{
			"misc.test": {
				Name:        "misc.test",
				Description: "General Usenet testing",
			},
			"alt.rfc-writers.recovery": {
				Name:        "alt.rfc-writers.recovery",
				Description: "RFC Writers Recovery",
			},
		}
		assert.Equal(t, expected, list)
	})

	t.Run("returns a list (with wildmat)", func(t *testing.T) {
		handler := func(t *testing.T, c net.Conn, cmd string, params []string) {
			assert.Equal(t, "list", cmd)
			assert.Equal(t, []string{"NEWSGROUPS", "*"}, params)
			writeLines(
				c,
				"215 list",
				"misc.test General Usenet testing",
				"alt.rfc-writers.recovery RFC Writers Recovery",
				".",
			)
		}

		server, client := getServerAndClient(t, handler)
		defer server.Close()

		list, err := client.ListNewsgroups("*")
		assert.Nil(t, err)

		expected := map[string]ListNewsgroup{
			"misc.test": {
				Name:        "misc.test",
				Description: "General Usenet testing",
			},
			"alt.rfc-writers.recovery": {
				Name:        "alt.rfc-writers.recovery",
				Description: "RFC Writers Recovery",
			},
		}
		assert.Equal(t, expected, list)
	})
}
