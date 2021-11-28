// Package service provides urx service.
package service

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func init() {
	// sets a seed for generating aliases
	rand.Seed(time.Now().UnixNano())
}

// Link is entity that connects URL and its aliases.
type Link struct {
	ID      string
	URL     string
	Aliases []string
}

const (
	// aliasAlphabet is a string that contains all possible characters
	// for URL alias.
	aliasAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	// aliasLength is the length of generated alias.
	aliasLength = 8
)

// NewLink creates and returns a new Link instance with generated alias.
func NewLink(URL string) Link {
	alias := make([]byte, aliasLength)
	for i := range alias {
		alias[i] = aliasAlphabet[rand.Intn(len(aliasAlphabet))]
	}

	return Link{ID: uuid.New().String(), URL: URL, Aliases: []string{string(alias)}}
}
