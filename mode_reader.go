package nntpclient

// ModeReader toggles the connection mode to "reader".
func (c *Client) ModeReader() error {
	code, message, err := c.sendCommand("MODE READER")
	if err != nil {
		return err
	}

	switch code {
	case 200:
		c.CanPost = true
	case 201:
		c.CanPost = false
	case 502:
		return ErrReadingUnavailable

	default:
		return UnexpectedError(code, message)
	}

	return nil
}
