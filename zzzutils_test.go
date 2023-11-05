package nntpclient

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
)

func getServerAndClient(t *testing.T, handler commandHandler) (*TestServer, *Client) {
	server, err := NewTestServer(t, handler)
	require.Nil(t, err)

	client, err := NewWithPort(server.Host, server.Port)
	require.Nil(t, err)

	err = client.Connect()
	require.Nil(t, err)

	return server, client
}

type commandHandler func(*testing.T, net.Conn, string, []string)

type TestServer struct {
	listener net.Listener

	handler commandHandler
	t       *testing.T

	Host string
	Port int
}

func NewTestServer(t *testing.T, handler commandHandler) (*TestServer, error) {
	server := &TestServer{t: t, handler: handler}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}

	server.listener = listener
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	server.Host = host
	server.Port = cast.ToInt(port)

	// Listen for future connections in the background.
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					continue
				}
				panic(err)
			}

			// A connection has been received, let's write to it and then
			// read it for commands.
			go func(c net.Conn) {
				writeLines(c, "200 welcome")
				server.router(c)
			}(conn)
		}
	}()

	return server, nil
}

func (s *TestServer) Close() error {
	return s.listener.Close()
}

func (s *TestServer) router(c net.Conn) {
	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		commandLine := scanner.Text()
		parts := strings.Fields(commandLine)
		name := strings.ToLower(parts[0])
		params := parts[1:]

		s.handler(s.t, c, name, params)
	}
}

func writeLines(c net.Conn, lines ...string) {
	for _, line := range lines {
		io.WriteString(c, line+"\r\n")
	}
}
