// Package service provides service for shortening URLs.
// Shortened URL is called URX.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"

	"urx/internal/config"
)

// LinkRepo is link repository.
type LinkRepo interface {
	Save(ctx context.Context, link Link) error
	FindByURL(ctx context.Context, URL string) (Link, error)
	FindByURX(ctx context.Context, URX string) (Link, error)
}

// Service is URL shortening service.
type Service struct {
	cfg config.Service
	r   LinkRepo
}

// NewService creates and returns a new Service instance.
func NewService(r LinkRepo) *Service {
	return &Service{cfg: config.Get().Service, r: r}
}

const RegexpURLPattern = `^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)+[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`

var (
	// ErrInvalidURL is returned when invalid URL was provided.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrLinkNotFound is returned when link was not found in store.
	ErrLinkNotFound = errors.New("link not found")
	// ErrGeneratingURX is returned when error occurred while generating URX.
	ErrGeneratingURX = errors.New("couldn't generate URX")
)

// Shorten shortens provided URL.
func (s *Service) Shorten(ctx context.Context, URL string) (URX string, err error) {
	if isValidURL, err := regexp.MatchString(RegexpURLPattern, URL); !isValidURL || err != nil {
		return "", ErrInvalidURL
	}

	if l, err := s.r.FindByURL(ctx, URL); err == nil {
		return fmt.Sprintf("%s/%s", s.cfg.Domain, l.URX), nil
	}

	if URX, err = s.generateURX(ctx); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.cfg.Domain, URX), s.r.Save(ctx, Link{URL: URL, URX: URX})
}

// generateURX generates random URX.
func (s *Service) generateURX(ctx context.Context) (urx string, err error) {
	bytes := make([]byte, 4)

	i := 0
	for {
		if _, err = rand.Read(bytes); err != nil {
			return "", err
		}
		urx = fmt.Sprintf("%x", bytes)

		if _, err = s.r.FindByURX(ctx, urx); err == ErrLinkNotFound {
			break
		}

		if i++; i >= 10 {
			return "", ErrGeneratingURX
		}
	}

	return urx, nil
}
