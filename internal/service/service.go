// Package service provides service for shortening URLs.
// Shortened URL is called URX.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"urx/internal/config"
)

var (
	// ErrInvalidURL is returned when invalid URL was provided.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrLinkNotFound is returned when link was not found in store.
	ErrLinkNotFound = errors.New("link not found")
	// ErrInvalidAlias is returned when invalid alias was provided.
	ErrInvalidAlias = errors.New("cannot use this alias")
)

// LinkRepo is link repository interface.
type LinkRepo interface {
	Save(ctx context.Context, l Link) error
	FindByURL(ctx context.Context, URL string) (Link, error)
	FindByAlias(ctx context.Context, alias string) (Link, error)
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
func (s *Service) Shorten(ctx context.Context, URL string) (URX string, err error) {
	if _, err = url.ParseRequestURI(URL); err != nil {
		return "", ErrInvalidURL
	}

	l, err := s.r.FindByURL(ctx, URL)
	if err == ErrLinkNotFound {
		l = NewLink(URL)
	} else if err != nil {
		return "", err
	}

	_, err = s.r.FindByAlias(ctx, l.Alias)
	if err == nil {
		return "", ErrInvalidAlias
	}
	if err != ErrLinkNotFound {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.cfg.Domain, l.Alias), s.r.Save(ctx, l)
}

// FindURL finds URL by alias.
func (s *Service) FindURL(ctx context.Context, alias string) (URL string, err error) {
	l, err := s.r.FindByAlias(ctx, alias)

	return l.URL, err
}
