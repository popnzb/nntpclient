package nntpclient

// Last sets the selected article to the most recent article in the
// selected group.
func (c *Client) Last() error {
	code, message, err := c.sendCommand("LAST")
	if err != nil {
		return err
	}

	switch code {
	case 412:
		return ErrNoGroupSelected
	case 420:
		return ErrCurrentArticleNumInvalid
	case 422:
		return ErrNoPrevArticle
	}

	if code != 223 {
		return UnexpectedError(code, message)
	}

	return nil
}
