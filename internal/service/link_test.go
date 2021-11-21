package service

import (
	"regexp"
	"testing"

	"github.com/google/uuid"

	"github.com/matryer/is"
)

func TestNewLink(t *testing.T) {
	is := is.New(t)
	l := NewLink("x.xx")

	_, err := uuid.Parse(l.ID)
	is.NoErr(err)

	is.Equal("x.xx", l.URL)

	matched, err := regexp.MatchString(GeneratedAliasRegExp, l.Aliases[0])
	if err != nil || !matched {
		t.Errorf("invalid alias: %v", l.Aliases[0])
	}
}
