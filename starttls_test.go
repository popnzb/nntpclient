package nntpclient

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StartTLS(t *testing.T) {
	t.Run("handles bad command response", func(t *testing.T) {
		client := Client{
			host: "example.com",
			conn: responseConn{response: &singleLineReader{line: "bad response\r\n"}},
		}

		err := client.StartTLS(nil)
		assert.ErrorContains(t, err, "could not process response code")
	})

	t.Run("errors for bad upgrade", func(t *testing.T) {
		client := Client{
			host: "example.com",
			conn: responseConn{response: &fullResponseReader{
				initial: "382 starttls",
				payload: []string{
					"111 date 20231112130000",
				},
			}},
		}

		err := client.StartTLS(nil)
		assert.ErrorContains(t, err, "first record does not look like a TLS")
	})

	t.Run("handles unexpected response code", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.Nil(t, err)
		defer listener.Close()
		//host, port, _ := net.SplitHostPort(listener.Addr().String())

		go func() {
			conn, _ := listener.Accept()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "STARTTLS") {
					io.WriteString(conn, "500 boom\r\n")
					return
				}
			}
		}()

		conn, err := net.Dial("tcp", listener.Addr().String())
		require.Nil(t, err)
		defer conn.Close()

		client := Client{conn: conn}

		err = client.StartTLS(nil)
		assert.ErrorContains(t, err, "unexpected response code: 500 (boom)")
	})

	t.Run("does upgrade", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.Nil(t, err)
		defer listener.Close()

		go func() {
			conn, _ := listener.Accept()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := scanner.Bytes()
				if strings.HasPrefix(string(line), "STARTTLS") {
					io.WriteString(conn, "382 do upgrade\r\n")
					conn = tls.Server(conn, &tls.Config{
						Certificates: []tls.Certificate{fakeCert},
					})
					// We need to redefine the scanner to read the upgraded
					// connection. Otherwise, we will read "garbage" going forward.
					scanner = bufio.NewScanner(conn)
					continue
				}
				if strings.HasPrefix(string(line), "DATE") {
					io.WriteString(conn, "111 20231112130000\r\n")
				}
			}
		}()

		conn, err := net.Dial("tcp", listener.Addr().String())
		require.Nil(t, err)
		defer conn.Close()

		client := Client{conn: conn}

		err = client.StartTLS(&tls.Config{InsecureSkipVerify: true})
		assert.Nil(t, err)
	})
}
