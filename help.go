package nntpclient

// Help retrieves the server help page for the support capabilities.
func (c *Client) Help() (string, error) {
	code, message, err := c.sendCommand("HELP")
	if err != nil {
		return "", err
	}
	if code != 100 {
		return "", UnexpectedError(code, message)
	}

	body, err := c.readBody()
	if err != nil {
		return "", err
	}

	return string(body), nil
}
