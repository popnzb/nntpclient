package nntpclient

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/textproto"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	c, err := New("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.1", c.host)
	assert.Equal(t, 119, c.port)
	assert.Nil(t, c.tlsConfig)
}

func Test_NewTls(t *testing.T) {
	c, err := NewTls("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, c.tlsConfig.ServerName, "127.0.0.1")
	assert.Equal(t, 563, c.port)
	assert.NotNil(t, c.tlsConfig)
}

func Test__new(t *testing.T) {
	t.Run("builds a default client", func(t *testing.T) {
		c, err := _new("127.0.0.1", 119)
		assert.Nil(t, err)
		assert.Equal(t, "127.0.0.1", c.host)
		assert.Equal(t, 119, c.port)
		assert.Nil(t, c.tlsConfig)
		assert.Equal(t, true, c.logger.Enabled(context.TODO(), slog.LevelInfo))
		assert.Equal(t, false, c.logger.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("processes options", func(t *testing.T) {
		dialer := &net.Dialer{}
		tlsConfig := &tls.Config{}

		var logs bytes.Buffer
		sink := logSink{buf: &logs}
		logger := slog.New(slog.NewJSONHandler(sink, nil))

		c, err := _new(
			"127.0.0.1",
			563,
			WithDialer(dialer),
			WithLogger(logger),
			WithTlsConfig(tlsConfig),
		)
		assert.Nil(t, err)
		assert.Equal(t, dialer, c.dialer)
		assert.Equal(t, tlsConfig, tlsConfig)

		c.logger.Info("test")
		assert.Contains(t, logs.String(), `"msg":"test"`)
	})
}

func Test_Connect(t *testing.T) {
	t.Run("handles non-tls dials with error", func(t *testing.T) {
		c := Client{
			host:   "127.0.0.1",
			port:   -1,
			dialer: &net.Dialer{Timeout: 0 * time.Nanosecond},
		}
		err := c.Connect()
		assert.ErrorContains(t, err, "dial tcp: address -1: invalid port")
	})

	t.Run("handles tls dials with error", func(t *testing.T) {
		c := Client{
			host:      "127.0.0.1",
			port:      -1,
			dialer:    &net.Dialer{Timeout: 0 * time.Nanosecond},
			tlsConfig: &tls.Config{},
		}
		err := c.Connect()
		assert.ErrorContains(t, err, "dial tcp: address -1: invalid port")
	})

	t.Run("handles a bad initial response", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.Nil(t, err)
		go func() {
			conn, _ := listener.Accept()
			io.WriteString(conn, "500 bad initial response\r\n")
		}()
		t.Cleanup(func() {
			listener.Close()
		})

		host, port, _ := net.SplitHostPort(listener.Addr().String())
		c := Client{
			host:   host,
			port:   cast.ToInt(port),
			dialer: &net.Dialer{},
		}

		err = c.Connect()
		assert.ErrorContains(t, err, "connection failure (code 500): bad initial response")
	})

	t.Run("handles a good connection", func(t *testing.T) {
		listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
			Certificates: []tls.Certificate{fakeCert},
		})
		require.Nil(t, err)
		go func() {
			conn, _ := listener.Accept()
			io.WriteString(conn, "200 welcome\r\n")
		}()
		t.Cleanup(func() {
			listener.Close()
		})

		host, port, _ := net.SplitHostPort(listener.Addr().String())
		c := Client{
			host:      host,
			port:      cast.ToInt(port),
			dialer:    &net.Dialer{},
			tlsConfig: &tls.Config{InsecureSkipVerify: true},
		}

		err = c.Connect()
		assert.Nil(t, err)
		assert.Equal(t, true, c.CanPost)
	})
}

func Test_sendCommand(t *testing.T) {
	t.Run("handles write error", func(t *testing.T) {
		c := Client{
			conn: &errWriterConn{},
		}

		code, message, err := c.sendCommand("whatever")
		assert.Equal(t, -1, code)
		assert.Equal(t, "", message)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("handles bad initial response", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &eofReader{}},
		}

		code, message, err := c.sendCommand("whatever")
		assert.Equal(t, -1, code)
		assert.Equal(t, "", message)
		assert.ErrorContains(t, err, "EOF")
	})

	t.Run("handles bad integer code", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "20x broken\r\n"}},
		}

		code, message, err := c.sendCommand("whatever")
		assert.Equal(t, -1, code)
		assert.Equal(t, "", message)
		assert.ErrorContains(t, err, "could not process response code:")
	})

	t.Run("handles success", func(t *testing.T) {
		c := Client{
			conn: responseConn{response: &singleLineReader{line: "200 ok\r\n"}},
		}

		code, message, err := c.sendCommand("whatever")
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Equal(t, "ok", message)
	})
}

func Test_readInitialResponse(t *testing.T) {
	t.Run("handles error from single line reader", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&eofReader{}),
		}
		c := Client{currentResponse: res}

		code, message, err := c.readInitialResponse()
		assert.Equal(t, -1, code)
		assert.Equal(t, "", message)
		assert.ErrorContains(t, err, "EOF")
	})

	t.Run("returns error for non-success code", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&singleLineReader{line: "404 missing\r\n"}),
		}
		c := Client{
			conn:            &net.TCPConn{},
			currentResponse: res,
		}

		code, message, err := c.readInitialResponse()
		assert.Equal(t, 404, code)
		assert.Equal(t, "missing", message)
		assert.ErrorContains(t, err, "connection failure (code 404): missing")
	})

	t.Run("returns success", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&singleLineReader{line: "200 ok\r\n"}),
		}
		c := Client{currentResponse: res}

		code, message, err := c.readInitialResponse()
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Equal(t, "ok", message)
	})
}

func Test_readSingleLineResponse(t *testing.T) {
	t.Run("returns errors", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&eofReader{}),
		}
		c := Client{currentResponse: res}

		line, err := c.readSingleLineResponse()
		assert.Equal(t, "", line)
		assert.ErrorContains(t, err, "EOF")
	})

	t.Run("reads actual response", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&singleLineReader{line: "200 ok\r\n"}),
		}
		c := Client{currentResponse: res}

		line, err := c.readSingleLineResponse()
		assert.Nil(t, err)
		assert.Equal(t, "200 ok\r\n", line)
	})
}

func Test_readHeaders(t *testing.T) {
	t.Run("returns error for EOF", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&eofReader{}),
		}
		c := Client{currentResponse: res}

		headers, err := c.readHeaders()
		assert.Nil(t, headers)
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("return error for bad read", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&boomReader{}),
		}
		c := Client{currentResponse: res}

		headers, err := c.readHeaders()
		assert.Nil(t, headers)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("reads unfolded standard headers", func(t *testing.T) {
		// i.e. a header block from an `ARTICLE` response
		reader := &multiLineReader{
			lines: []string{
				"foo: foo",
				"bar: bar",
				"baz: baz",
				"",
			},
		}
		res := &Response{
			bufferedReader: bufio.NewReader(reader),
		}
		c := Client{currentResponse: res}

		headers, err := c.readHeaders()
		assert.Nil(t, err)

		expected := textproto.MIMEHeader{
			"Foo": {"foo"},
			"Bar": {"bar"},
			"Baz": {"baz"},
		}
		assert.Equal(t, expected, headers)
	})

	t.Run("reads unfolded dot terminated headers", func(t *testing.T) {
		// i.e. a header block from an `HEAD` response
		reader := &multiLineReader{
			lines: []string{
				"foo: foo",
				"bar: bar",
				"baz: baz",
				".",
			},
		}
		res := &Response{
			bufferedReader: bufio.NewReader(reader),
		}
		c := Client{currentResponse: res}

		headers, err := c.readHeaders()
		assert.Nil(t, err)

		expected := textproto.MIMEHeader{
			"Foo": {"foo"},
			"Bar": {"bar"},
			"Baz": {"baz"},
		}
		assert.Equal(t, expected, headers)
	})

	t.Run("reads folded", func(t *testing.T) {
		reader := &multiLineReader{
			lines: []string{
				"foo: foo;",
				" foo;",
				" foo;",
				" foo",
				"bar: bar;",
				"\tbar",
				"baz: baz",
				"",
			},
		}
		res := &Response{
			bufferedReader: bufio.NewReader(reader),
		}
		c := Client{currentResponse: res}

		headers, err := c.readHeaders()
		assert.Nil(t, err)

		expected := textproto.MIMEHeader{
			"Foo": {"foo; foo; foo; foo"},
			"Bar": {"bar;\tbar"},
			"Baz": {"baz"},
		}
		assert.Equal(t, expected, headers)
	})

	t.Run("returns error for unexpted folded header", func(t *testing.T) {
		reader := &multiLineReader{
			lines: []string{"\tfoo: foo"},
		}
		res := &Response{
			bufferedReader: bufio.NewReader(reader),
		}
		c := Client{currentResponse: res}

		headers, err := c.readHeaders()
		assert.Nil(t, headers)
		assert.ErrorContains(t, err, "malformed headers, found")
	})
}

func Test_readBody(t *testing.T) {
	t.Run("returns error for EOF", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&eofReader{}),
		}
		c := Client{currentResponse: res}

		var body bytes.Buffer
		err := c.readBody(&body)
		assert.Equal(t, "eofReader", body.String())
		assert.Equal(t, true, errors.Is(err, NntpError))
		assert.ErrorContains(t, err, "unexpected end of response")
	})

	t.Run("returns error for bad read", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&boomReader{}),
		}
		c := Client{
			currentResponse: res,
		}

		var body bytes.Buffer
		err := c.readBody(&body)
		assert.Equal(t, "", body.String())
		assert.Error(t, err)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("returns expected bytes", func(t *testing.T) {
		res := &Response{
			bufferedReader: bufio.NewReader(&multiLineReader{
				lines: []string{"success", "."},
			}),
		}
		c := Client{
			currentResponse: res,
		}

		var body bytes.Buffer
		err := c.readBody(&body)
		assert.Nil(t, err)
		assert.Equal(t, "success\r\n", body.String())
	})
}
