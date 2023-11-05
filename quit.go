package nntpclient

// Close is an alias for the Quit method.
func (c *Client) Close() error {
	return c.Quit()
}

// Quit sends a standard `QUIT` to the remote server and terminates
// the connection.
func (c *Client) Quit() error {
	_, _, err := c.sendCommand("QUIT")
	if err != nil {
		return err
	}

	c.conn.Close()
	return nil
}
