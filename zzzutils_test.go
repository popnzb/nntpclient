package nntpclient

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

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

/** Logger Mocks */

var NilLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type logSink struct {
	buf *bytes.Buffer
}

func (ls logSink) Write(p []byte) (int, error) {
	return ls.buf.Write(p)
}

/** Reader Mocks */

// eofReader writes "eofReader" to the buffer and returns [io.EOF]
// on every read.
type eofReader struct{}

func (er *eofReader) Read(buf []byte) (int, error) {
	result := "eofReader"
	copy(buf, result)
	return len(result), io.EOF
}

// boomReader returns an error, "boom", for every read.
type boomReader struct{}

func (b *boomReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

// multiLineReader writes all lines, with automatic `\r\n` termination,
// on every read.
type multiLineReader struct {
	lines []string
}

func (mlr *multiLineReader) Read(buf []byte) (int, error) {
	bytesRead := 0
	var output bytes.Buffer
	for _, line := range mlr.lines {
		toRead := line + "\r\n"
		x, _ := output.Write([]byte(toRead))
		bytesRead += x
	}
	copy(buf, output.Bytes())
	return bytesRead, nil
}

// singleLineReader always returns a single line for any given read. The
// line should be terminated with `\r\n`.
type singleLineReader struct {
	line string
}

func (slr *singleLineReader) Read(buf []byte) (int, error) {
	copy(buf, slr.line)
	return len(slr.line), nil
}

// badBodyReader returns a valid response line on the first read, but
// any subsequent read will return an [io.EOF]. The initial string does
// not need to be terminated with `\r\n`.
type badBodyReader struct {
	sentInitial bool
	initial     string
}

func (bbr *badBodyReader) Read(buf []byte) (int, error) {
	if bbr.sentInitial == false {
		count := copy(buf, bbr.initial+"\r\n")
		bbr.sentInitial = true
		return count, nil
	}

	count := copy(buf, "bad body\r\n")
	return count, io.EOF
}

// fullResponseReader writes initial on the first read. On subsequent reads,
// it writes all of payload to the buffer. All strings are automatically
// terminated with `\r\n`, but the payload should end with a single `"."`
// string.
type fullResponseReader struct {
	sentInitial bool
	initial     string
	payload     []string
}

func (frr *fullResponseReader) Read(buf []byte) (int, error) {
	if frr.sentInitial == false {
		count := copy(buf, frr.initial+"\r\n")
		frr.sentInitial = true
		return count, nil
	}

	var toRead bytes.Buffer
	for _, line := range frr.payload {
		toRead.WriteString(line + "\r\n")
	}
	copy(buf, toRead.Bytes())
	return len(toRead.Bytes()), nil
}

/** Connection Mocks */

type errWriterConn struct{}

func (er errWriterConn) Read([]byte) (int, error) {
	return 0, nil
}

func (er errWriterConn) Close() error {
	return nil
}

func (er errWriterConn) LocalAddr() net.Addr {
	return nil
}

func (er errWriterConn) RemoteAddr() net.Addr {
	return nil
}

func (er errWriterConn) SetDeadline(time.Time) error {
	return nil
}

func (er errWriterConn) SetReadDeadline(time.Time) error {
	return nil
}

func (er errWriterConn) SetWriteDeadline(time.Time) error {
	return nil
}

func (er errWriterConn) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}

type responseConn struct {
	response io.Reader
}

func (rc responseConn) Read(buf []byte) (int, error) {
	return rc.response.Read(buf)
}

func (rc responseConn) Close() error {
	return nil
}

func (rc responseConn) LocalAddr() net.Addr {
	return nil
}

func (rc responseConn) RemoteAddr() net.Addr {
	return nil
}

func (rc responseConn) SetDeadline(time.Time) error {
	return nil
}

func (rc responseConn) SetReadDeadline(time.Time) error {
	return nil
}

func (rc responseConn) SetWriteDeadline(time.Time) error {
	return nil
}

func (rc responseConn) Write([]byte) (int, error) {
	return 0, nil
}

/** Fake Certificate */
// This cert is pulled directly from the X509KeyPair example in the
// crypto.tls package.
var certPem = []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)
var keyPem = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)
var fakeCert, _ = tls.X509KeyPair(certPem, keyPem)
