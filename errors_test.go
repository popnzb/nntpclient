package nntpclient

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SpecificErrorsWrapBase(t *testing.T) {
	assert.Equal(t, true, errors.Is(ErrCurrentArticleNumInvalid, NntpError))
	assert.Equal(t, true, errors.Is(ErrNoArticleWithId, NntpError))
	assert.Equal(t, true, errors.Is(ErrNoArticleWithNum, NntpError))
	assert.Equal(t, true, errors.Is(ErrNoGroupSelected, NntpError))
	assert.Equal(t, true, errors.Is(ErrNoNextArticle, NntpError))
	assert.Equal(t, true, errors.Is(ErrNoPrevArticle, NntpError))
	assert.Equal(t, true, errors.Is(ErrReadingUnavailable, NntpError))
	assert.Equal(t, true, errors.Is(ErrNoSuchGroup, NntpError))
}

func Test_AuthError(t *testing.T) {
	err := AuthError(111, "foo")
	assert.Equal(t, true, errors.Is(err, NntpError))
	fmt.Printf("%+v", err)
}

func Test_UnexpectedError(t *testing.T) {
	err := UnexpectedError(111, "foo")
	assert.Equal(t, true, errors.Is(err, NntpError))
}
