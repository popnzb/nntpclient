package nntpclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

const lineTerminatorByte = 0x0a // "\n"

// Client is a simple [NNTP](https://datatracker.ietf.org/doc/html/rfc3977)
// client. Client instances should be created with [New], [NewTls], or
// [NewWithPort] (this one being the most flexible).
type Client struct {
	host      string
	port      int
	dialer    *net.Dialer
	tlsConfig *tls.Config
	conn      net.Conn

	logger *slog.Logger

	currentResponse *Response

	// CanPost indicates if the server will allow the client to post articles.
	// Useful in the future when posting is supported by the client.
	CanPost bool
}

type Option func(client *Client)

// New creates a new [Client] instance that connects to the given `host`
// according to the given options. Instances created with this method will
// _always_ try to connect on port 119. Thus, it is not compatible with
// the [WithTlsConfig] options.
func New(host string, opts ...Option) (*Client, error) {
	return NewWithPort(host, 119, opts...)
}

// NewTls creates a new [Client] instance that connects to the given `host`
// on port 563 utilizing a basic TLS configuration that sets the `ServerName`
// option to `host`. This should be good enough for most standard NNTP servers
// that support TLS on the standard port.
func NewTls(host string, opts ...Option) (*Client, error) {
	opts = append(opts, WithTlsConfig(&tls.Config{ServerName: host}))
	return NewWithPort(host, 563, opts...)
}

// NewWithPort is the most generic method for creating a new [Client]
// instance. It supports all options and allows flexibility in defining the
// destination port. If no TLS configuration is provided, via [WithTlsConfig],
// then the connection will be in plain text.
func NewWithPort(host string, port int, opts ...Option) (*Client, error) {
	return _new(host, port, opts...)
}

// WithDialer allows defining the dialer that will be used to establish
// the connection with the remote server.
func WithDialer(dialer *net.Dialer) Option {
	return func(client *Client) {
		client.dialer = dialer
	}
}

// WithLogger allows defining the logger instance that will be used when
// logging messages. The default logger logs at the "info" level to the
// stdout stream.
func WithLogger(logger *slog.Logger) Option {
	return func(client *Client) {
		client.logger = logger
	}
}

// WithTlsConfig allows defining the TLS configuration to be used when
// establishing a connection to a TLS enabled port.
func WithTlsConfig(config *tls.Config) Option {
	return func(client *Client) {
		client.tlsConfig = config
	}
}

func _new(host string, port int, opts ...Option) (*Client, error) {
	logOptions := &slog.HandlerOptions{Level: slog.LevelInfo}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, logOptions))
	client := &Client{
		host: host,
		port: port,
		dialer: &net.Dialer{
			KeepAlive: 0,
		},
		logger: logger,
	}

	for _, opt := range opts {
		opt(client)
	}

	client.logger = client.logger.WithGroup("nntpclient")

	return client, nil
}

// Connect establishes a connection to the server. This method must be invoked
// once prior to any command methods.
func (c *Client) Connect() error {
	// TODO: add a c.connected property and check it here
	address := net.JoinHostPort(c.host, fmt.Sprint(c.port))

	switch {
	case c.tlsConfig == nil:
		conn, err := c.dialer.Dial("tcp", address)
		if err != nil {
			return err
		}
		c.conn = conn
	case c.tlsConfig != nil:
		conn, err := tls.DialWithDialer(c.dialer, "tcp", address, c.tlsConfig)
		if err != nil {
			return err
		}
		c.conn = conn
	}

	c.currentResponse = NewResponse(c.conn)
	code, _, err := c.readInitialResponse()
	if err != nil {
		return err
	}

	if code == 200 {
		c.CanPost = true
	}

	return nil
}

// sendCommand writes the provided command to the server, processes the
// initial response line, and returns the information processed from that
// line. If a command returns more data than a single response line, the
// [readHeaders] and [readBody] methods should be used subsequent to this
// method.
func (c *Client) sendCommand(command string) (code int, message string, err error) {
	_, err = fmt.Fprintf(c.conn, "%s\r\n", command)
	if err != nil {
		return -1, "", err
	}

	c.currentResponse = NewResponse(c.conn)

	line, err := c.readSingleLineResponse()
	if err != nil {
		return -1, "", err
	}
	code, err = strconv.Atoi(line[0:3])
	if err != nil {
		return -1, "", fmt.Errorf("could not process response code: %v", err)
	}

	return code, strings.TrimSpace(line[3:]), nil
}

// readInitialResponse reads the response sent by the server upon initial
// connection. The spec requires a specific code to start the response. Anything
// after the code in the initial response is arbitrarily set by the
// server. This arbitrary string will be returned for completeness’s sake.
func (c *Client) readInitialResponse() (int, string, error) {
	line, err := c.readSingleLineResponse()
	if err != nil {
		return -1, "", err
	}

	code := cast.ToInt(line[0:3])
	message := strings.TrimSpace(line[4:])

	if code != 200 && code != 201 {
		defer c.conn.Close()
		err := fmt.Errorf("connection failure (code %d): %s", code, message)
		return code, message, err
	}

	return code, message, nil
}

// readSingleLineResponse is used to read a standard single line response from
// the server. See https://datatracker.ietf.org/doc/html/rfc3977#section-3.1
// for information about standard single line responses.
//
// The unprocessed line will be returned.
func (c *Client) readSingleLineResponse() (string, error) {
	readBytes, err := c.currentResponse.ReadBytes(lineTerminatorByte)
	if err != nil {
		return "", err
	}
	return string(readBytes), nil
}

// readHeaders is used to read a headers, or headers-like, block after issuing
// a command.
func (c *Client) readHeaders() (textproto.MIMEHeader, error) {
	headers, _, err := ReadHeaders(c.currentResponse)
	return headers, err
}

// readBody is used to read a body, or body-like, block from a response after
// issuing a command. The body is NOT processed. It is instead written to the
// passed in [io.Writer]. This allows for processing of bodies according to
// their content by a client.
func (c *Client) readBody(writer io.Writer) error {
	return ReadBody(c.currentResponse, writer)
}

// ReadHeaders parses a set of bytes with the expectation that they start
// with what look like header lines terminated by either an empty line,
// in the case of a header block at the top of an article, or the message
// termination line (`.\r\n`) as when reading a `HEAD` command response.
//
// The result of this method, in the success case, is a map of headers and
// an integer representing the offset of the last read byte, e.g. the start
// of the body in an article. Otherwise, and error is returned and the other
// values may be incomplete or incorrect.
//
// If an end of file is reached before a header block termination line then
// the [io.EOF] error will be returned along with the last read offset.
func ReadHeaders(reader MultibyteReader) (textproto.MIMEHeader, int, error) {
	offset := 0
	result := make(textproto.MIMEHeader)

	// lastHeader is the most recently found header. We need this if the header
	// value has been folded.
	lastHeader := ""
	// endOfHeaders is the empty line between a header block and a body block.
	endOfHeaders := []byte("\r\n")
	// endOfResponse is the single dot that follows an informational command like "HEAD".
	endOfResponse := []byte(".\r\n")
	for {
		readBytes, err := reader.ReadBytes(lineTerminatorByte)
		offset += len(readBytes)

		if err != nil {
			if err == io.EOF {
				return nil, offset, ErrUnexpectedEOF
			}
			return nil, offset, err
		}

		if bytes.Equal(readBytes, endOfHeaders) || bytes.Equal(readBytes, endOfResponse) {
			break
		}

		leadingByte := readBytes[0]
		if leadingByte == 0x20 || leadingByte == 0x09 {
			// Line starts with a space character or a tab character.
			// Therefore, it must be a folded value.
			if lastHeader == "" {
				return nil, offset, errors.New("malformed headers, found folded value without name")
			}

			values := result.Values(lastHeader)
			lastValue := values[len(values)-1]
			lastValue = lastValue + string(readBytes[:len(readBytes)-2])
			values[len(values)-1] = lastValue
			continue
		}

		colonIndex := bytes.IndexByte(readBytes, 0x3a)
		name := string(readBytes[0:colonIndex])
		// Omit the leading ": ", and trailing "\r\n"
		value := string(readBytes[colonIndex+2 : len(readBytes)-2])
		result.Add(name, value)
		lastHeader = name
	}

	return result, offset, nil
}

// ReadBody is used to read a body, or body-like, block of bytes provided
// by the passed in reader. The body is not processed in any way, but is
// instead written to the supplied writer. This allows for processing of
// bodies according to their content by a client. Reading stops when the
// termination line (`.\r\n`) is encountered.
//
// If an end of file is encountered before the termination line, the [io.EOF]
// error will be returned. In this case, it is not advisable to trust the
// bytes written to the writer.
func ReadBody(reader MultibyteReader, writer io.Writer) error {
	endOfBody := []byte(".\r\n")
	for {
		readBytes, err := reader.ReadBytes(lineTerminatorByte)
		if err != nil {
			if err == io.EOF {
				writer.Write(readBytes)
				return ErrUnexpectedEOF
			}
			return err
		}

		if bytes.Equal(readBytes, endOfBody) {
			break
		}

		writer.Write(readBytes)
	}

	return nil
}
