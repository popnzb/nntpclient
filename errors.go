package nntpclient

import (
	"errors"
	"fmt"
)

// NntpError is the base error for all errors that derive from issuing
// commands to the server. Any such errors can be checked with [errors.Is].
var NntpError = errors.New("nntp error")

/** Errors that stem directly from the spec. */

var ErrCurrentArticleNumInvalid = fmt.Errorf("current article number is invalid: %w", NntpError)
var ErrNoArticleWithId = fmt.Errorf("no article with that message-id: %w", NntpError)
var ErrNoArticleWithNum = fmt.Errorf("no article with that number: %w", NntpError)
var ErrNoGroupSelected = fmt.Errorf("no newsgroup selected: %w", NntpError)
var ErrNoNextArticle = fmt.Errorf("no next article in this group: %w", NntpError)
var ErrNoPrevArticle = fmt.Errorf("no previous article in this group: %w", NntpError)
var ErrReadingUnavailable = fmt.Errorf("reading service permanently unavailable: %w", NntpError)
var ErrNoSuchGroup = fmt.Errorf("no such newsgroup found: %w", NntpError)

/** Library specific errors that are still NNTP derived. */

var ErrUnexpectedEOF = fmt.Errorf("unexpected end of response: %w", NntpError)

func AuthError(code int, message string) error {
	return fmt.Errorf("auth failed with code: %d (%s): %w", code, message, NntpError)
}

func UnexpectedError(code int, message string) error {
	return fmt.Errorf("unexpected response code: %d (%s): %w", code, message, NntpError)
}
