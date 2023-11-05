package nntpclient

import (
	"bufio"
	"bytes"
	"strings"
)

// Capabilities is a map of all capability labels to the label's possible
// arguments. For example, the `COMPRESS` label may have a list of possible
// arguments that looks like `[DEFLATE]`.
type Capabilities map[string][]string

// Capabilities gets a mapping of all supported labels to their possible
// arguments.
func (c *Client) Capabilities() (*Capabilities, error) {
	code, message, err := c.sendCommand("CAPABILITIES")
	if err != nil {
		return nil, err
	}
	c.logger.Debug("capabilities", "code", code, "message", message)

	if code != 101 {
		return nil, UnexpectedError(code, message)
	}

	bodyBytes, err := c.readBody()
	if err != nil {
		return nil, err
	}

	capabilities := make(Capabilities)
	scanner := bufio.NewScanner(bytes.NewReader(bodyBytes))
	for scanner.Scan() {
		scanLine := scanner.Text()
		parts := strings.Fields(scanLine)
		capabilities[parts[0]] = parts[1:]
	}

	return &capabilities, nil
}
