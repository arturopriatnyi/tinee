// Package service provides service for shortening URLs.
// Shortened URL is called URX.
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"urx/internal/config"
)

const (
	// URLRegExp is regular expression pattern for URL.
	URLRegExp = "^(?:http(s)?:\\/\\/)[\\w.-]+(?:\\.[\\w\\.-]+)+[\\w\\-\\._~:/?#[\\]@!\\$&'\\(\\)\\*\\+,;=.]+$"
	// GeneratedAliasRegExp is regular expression patter for generated aliases.
	GeneratedAliasRegExp = "^[a-zA-Z0-9]{8}$"
	// CustomAliasRegExp is regular expression pattern for custom aliases.
	CustomAliasRegExp = "^[a-zA-Z0-9]{4,}$"
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

// Service is URL shortening service.
type Service struct {
	cfg config.Service
	r   LinkRepo
}

// New creates and returns a new Service instance.
func New(cfg config.Service, r LinkRepo) *Service {
	return &Service{cfg: cfg, r: r}
}

// Shorten shortens provided URL.
func (s *Service) Shorten(ctx context.Context, URL, alias string) (URX string, err error) {
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
		return s.URX(link.Aliases[0]), nil
	}

	if err = s.ValidateAlias(alias); err != nil {
		return "", err
	}
	l, err := s.r.FindByAlias(ctx, alias)
	if err == ErrLinkNotFound {
		link.Aliases = append(link.Aliases, alias)
	} else if err == nil && link.ID != l.ID {
		return "", ErrInvalidAlias
	} else if err != nil {
		return "", err
	}

	return s.URX(alias), s.r.Save(ctx, link)
}

// URLByAlias returns URL by alias.
func (s *Service) URLByAlias(ctx context.Context, alias string) (URL string, err error) {
	l, err := s.r.FindByAlias(ctx, alias)

	return l.URL, err
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

// URX forms URX with provided alias.
func (s *Service) URX(alias string) string {
	return fmt.Sprintf("%s/%s", s.cfg.Domain, alias)
}

// ValidateURL validates URL.
func (s *Service) ValidateURL(URL string) error {
	if matched, err := regexp.MatchString(URLRegExp, URL); err != nil || !matched {
		return ErrInvalidURL
	}

	return nil
}

// ValidateAlias validates alias.
func (s *Service) ValidateAlias(alias string) error {
	if matched, err := regexp.MatchString(CustomAliasRegExp, alias); err != nil || !matched {
		return ErrInvalidAlias
	}

	return nil
}
