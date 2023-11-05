package nntpclient

// Next selects the next article in the selected group.
func (c *Client) Next() error {
	code, message, err := c.sendCommand("NEXT")
	if err != nil {
		return err
	}

	switch code {
	case 412:
		return ErrNoGroupSelected
	case 420:
		return ErrCurrentArticleNumInvalid
	case 421:
		return ErrNoNextArticle
	}

	if code != 223 {
		return UnexpectedError(code, message)
	}

	return nil
}
