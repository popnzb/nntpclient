package nntpclient

// Authenticate provides simple username and password authentication through
// the AUTHINFO extension (RFC 4643). The absence of an error indicates
// successful authentication.
func (c *Client) Authenticate(user string, pass string) error {
	code, message, err := c.sendCommand("AUTHINFO USER " + user)
	if err != nil {
		return err
	}
	if code != 381 {
		return UnexpectedError(code, message)
	}

	code, message, err = c.sendCommand("AUTHINFO PASS " + pass)
	if err != nil {
		return err
	}
	if code != 281 {
		return AuthError(code, message)
	}

	return nil
}
