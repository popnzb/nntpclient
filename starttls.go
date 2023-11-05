package nntpclient

import (
	"crypto/tls"
)

// StartTLS upgrades the current connection to a TLS protected one.
// See RFC 4642.
//
// Note: if a config is not provided, one with the `ServerName` set to the
// host will be used.
func (c *Client) StartTLS(config *tls.Config) error {
	// This implementation is based upon
	// https://cs.opensource.google/go/go/+/refs/tags/go1.21.3:src/net/smtp/smtp.go;l=154-166

	if config == nil {
		config = &tls.Config{ServerName: c.host}
	}

	code, message, err := c.sendCommand("STARTTLS")
	if err != nil {
		return err
	}
	if code != 382 {
		return UnexpectedError(code, message)
	}

	c.conn = tls.Client(c.conn, config)

	// Verify that the upgrade has worked. If we get an error, it's likely
	// a certificate error.
	_, err = c.Date()
	if err != nil {
		return err
	}

	return nil
}
