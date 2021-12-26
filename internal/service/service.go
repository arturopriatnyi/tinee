// Package service provides service for shortening URLs.
// Shortened URL is called tineeURL.
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"tinee/internal/config"
)

const (
	// URLRegExp is regular expression pattern for URL.
	URLRegExp = "^(?:http(s)?:\\/\\/)[\\w.-]+(?:\\.[\\w\\.-]+)+[\\w\\-\\._~:/?#[\\]@!\\$&'\\(\\)\\*\\+,;=.]+$"
	// GeneratedAliasRegExp is regular expression pattern for generated aliases.
	GeneratedAliasRegExp = "^[A-Za-z0-9]{8}$"
	// CustomAliasRegExp is regular expression pattern for custom aliases.
	CustomAliasRegExp = "^[A-Za-z0-9]{4,}$"
)

var (
	// ErrInvalidURL is returned when invalid URL was provided.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrInvalidAlias is returned when invalid alias was provided.
	ErrInvalidAlias = errors.New("invalid alias")
	// ErrLinkNotFound is returned when link was not found in store.
	ErrLinkNotFound = errors.New("link not found")
)

// LinkRepo is link repository interface.
type LinkRepo interface {
	Save(context.Context, Link) error
	FindByURL(context.Context, string) (Link, error)
	FindByAlias(context.Context, string) (Link, error)
}

// LinkCache is link cache interface.
type LinkCache interface {
	Set(ctx context.Context, alias string, l Link) error
	Get(ctx context.Context, alias string) (Link, error)
}

// Service is URL shortening service.
type Service struct {
	cfg config.Service
	r   LinkRepo
	c   LinkCache
}

// New creates and returns a new Service instance.
func New(cfg config.Service, r LinkRepo, c LinkCache) *Service {
	return &Service{cfg: cfg, r: r, c: c}
}

// Shorten shortens provided URL.
func (s *Service) Shorten(ctx context.Context, URL, alias string) (tineeURL string, err error) {
	if err = s.ValidateURL(URL); err != nil {
		return "", err
	}

	link, err := s.r.FindByURL(ctx, URL)
	if err == ErrLinkNotFound {
		link, err = s.CreateLink(ctx, URL)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	if alias == "" {
		return s.TineeURL(link.Aliases[0]), nil
	}

	if err = s.ValidateCustomAlias(alias); err != nil {
		return "", err
	}
	l, err := s.LinkByAlias(ctx, alias)
	if err == ErrLinkNotFound {
		link.Aliases = append(link.Aliases, alias)
	} else if err == nil && link.ID != l.ID {
		return "", ErrInvalidAlias
	} else if err != nil {
		return "", err
	}

	return s.TineeURL(alias), s.r.Save(ctx, link)
}

// LinkByAlias finds and returns a Link by alias.
func (s *Service) LinkByAlias(ctx context.Context, alias string) (l Link, err error) {
	if l, err = s.c.Get(ctx, alias); err == nil {
		return l, nil
	}

	if l, err = s.r.FindByAlias(ctx, alias); err == nil {
		_ = s.c.Set(ctx, alias, l)
	}

	return l, err
}

// CreateLink creates a Link with provided URL and generated alias.
func (s *Service) CreateLink(ctx context.Context, URL string) (l Link, err error) {
	l = NewLink(URL)
	if _, err = s.r.FindByAlias(ctx, l.Aliases[0]); err == nil {
		return Link{}, ErrInvalidAlias
	} else if err != nil && err != ErrLinkNotFound {
		return Link{}, err
	}

	return l, s.r.Save(ctx, l)
}

// TineeURL forms tineeURL with provided alias.
func (s *Service) TineeURL(alias string) string {
	return fmt.Sprintf("%s/%s", s.cfg.Domain, alias)
}

// ValidateURL validates URL.
func (s *Service) ValidateURL(URL string) error {
	if matched, err := regexp.MatchString(URLRegExp, URL); err != nil || !matched {
		return ErrInvalidURL
	}

	return nil
}

// ValidateCustomAlias validates custom alias.
func (s *Service) ValidateCustomAlias(alias string) error {
	if matched, err := regexp.MatchString(CustomAliasRegExp, alias); err != nil || !matched {
		return ErrInvalidAlias
	}

	return nil
}
