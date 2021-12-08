package service

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/matryer/is"

	"urx/internal/config"
)

type mockLinkRepo struct {
	save        func(context.Context, Link) error
	findByURL   func(context.Context, string) (Link, error)
	findByAlias func(context.Context, string) (Link, error)
}

func (r *mockLinkRepo) Save(ctx context.Context, link Link) error {
	return r.save(ctx, link)
}

func (r *mockLinkRepo) FindByURL(ctx context.Context, URL string) (Link, error) {
	return r.findByURL(ctx, URL)
}

func (r *mockLinkRepo) FindByAlias(ctx context.Context, alias string) (Link, error) {
	return r.findByAlias(ctx, alias)
}

type mockLinkCache struct {
	get func(context.Context, string) (Link, error)
	set func(context.Context, string, Link) error
}

func (c *mockLinkCache) Get(ctx context.Context, alias string) (Link, error) {
	return c.get(ctx, alias)
}

func (c *mockLinkCache) Set(ctx context.Context, alias string, l Link) error {
	return c.set(ctx, alias, l)
}

func TestNewService(t *testing.T) {
	is := is.New(t)
	cfg := config.Service{}
	r := &mockLinkRepo{}
	c := &mockLinkCache{}

	is.Equal(&Service{cfg: cfg, r: r, c: c}, New(cfg, r, c))
}

func TestService_Shorten(t *testing.T) {
	testcases := []struct {
		name   string
		r      *mockLinkRepo
		c      *mockLinkCache
		url    string
		alias  string
		expErr error
	}{
		{
			name: "URL is shortened with generated alias",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			expErr: nil,
		},
		{
			name: "URL is shortened with custom alias",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: nil,
		},
		{
			name:   "invalid URL",
			url:    "x.xx",
			expErr: ErrInvalidURL,
		},
		{
			name: "FindByURL unexpected error",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, errors.New("unexpected error")
				},
			},
			url:    "https://x.xx",
			expErr: errors.New("unexpected error"),
		},
		{
			name: "FindByAlias invalid alias error while creating link",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, nil
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			expErr: ErrInvalidAlias,
		},
		{
			name: "FindByAlias unexpected error while creating link",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, errors.New("unexpected error")
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			expErr: errors.New("unexpected error"),
		},
		{
			name: "invalid custom alias",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			alias:  "x",
			expErr: ErrInvalidAlias,
		},
		{
			name: "FindByAlias invalid alias error while adding custom alias",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					// "xxxx" alias is already taken
					if alias == "xxxx" {
						return Link{ID: "x-x-x-x"}, nil
					}

					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: ErrInvalidAlias,
		},
		{
			name: "FindByAlias unexpected error while adding custom alias",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					if alias == "xxxx" {
						return Link{}, errors.New("unexpected error")
					}

					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			s := New(config.Service{}, tc.r, tc.c)

			urx, err := s.Shorten(context.Background(), tc.url, tc.alias)

			is.Equal(tc.expErr, err)
			matched, err := regexp.MatchString(`/[a-zA-Z0-9]`, urx)
			if tc.expErr == nil && (err != nil || !matched) {
				t.Errorf("invalid URX: %s", urx)
			}
		})
	}
}

func TestService_LinkByAlias(t *testing.T) {
	testcases := []struct {
		name    string
		r       *mockLinkRepo
		c       *mockLinkCache
		urx     string
		expLink Link
		expErr  error
	}{
		{
			name: "link is found in repository",
			r: &mockLinkRepo{
				findByAlias: func(ctx context.Context, URX string) (Link, error) {
					return Link{URL: "https://x.xx"}, nil
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			urx:     "xxxxxxxx",
			expLink: Link{URL: "https://x.xx"},
		},
		{
			name: "link is found in cache",
			r: &mockLinkRepo{
				findByAlias: func(ctx context.Context, URX string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
			},
			c: &mockLinkCache{
				get: func(ctx context.Context, alias string) (Link, error) {
					return Link{URL: "https://x.xx"}, nil
				},
				set: func(ctx context.Context, alias string, l Link) error {
					return nil
				},
			},
			urx:     "xxxxxxxx",
			expLink: Link{URL: "https://x.xx"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			s := New(config.Service{}, tc.r, tc.c)

			l, err := s.LinkByAlias(context.Background(), tc.urx)

			is.Equal(tc.expErr, err)
			is.Equal(tc.expLink, l)
		})
	}
}

func TestService_CreateLink(t *testing.T) {
	testcases := []struct {
		name   string
		r      *mockLinkRepo
		url    string
		expErr error
	}{
		{
			name: "link is created",
			r: &mockLinkRepo{
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			url: "https://x.xx",
		},
		{
			name: "invalid alias error",
			r: &mockLinkRepo{
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, nil
				},
			},
			url:    "https://x.xx",
			expErr: ErrInvalidAlias,
		},
		{
			name: "unexpected error",
			r: &mockLinkRepo{
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					return Link{}, errors.New("unexpected error")
				},
			},
			url:    "https://x.xx",
			expErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			s := New(config.Service{}, tc.r, nil)

			l, err := s.CreateLink(context.Background(), tc.url)

			is.Equal(tc.expErr, err)
			if tc.expErr == nil && l.ID == "" {
				t.Errorf("invalid link: %v", l)
			}
		})
	}
}

func TestService_URX(t *testing.T) {
	is := is.New(t)
	s := New(config.Service{Domain: "urx.io"}, nil, nil)

	is.Equal("urx.io/xxxx", s.URX("xxxx"))
}

func TestService_ValidateURL(t *testing.T) {
	testcases := []struct {
		name   string
		url    string
		expErr error
	}{
		{
			name:   "url is valid with https",
			url:    "https://x.xx",
			expErr: nil,
		},
		{
			name:   "url is valid with http",
			url:    "http://x.xx",
			expErr: nil,
		},
		{
			name:   "url is not valid without schema",
			url:    "x.xx",
			expErr: ErrInvalidURL,
		},
		{
			name:   "url is not valid without domain",
			url:    "https://x",
			expErr: ErrInvalidURL,
		},
		{
			name:   "url is not valid without valid domain",
			url:    "http://x.x",
			expErr: ErrInvalidURL,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			s := New(config.Service{}, nil, nil)

			is.Equal(tc.expErr, s.ValidateURL(tc.url))
		})
	}
}

func TestService_ValidateCustomAlias(t *testing.T) {
	testcases := []struct {
		name   string
		alias  string
		expErr error
	}{
		{
			name:   "alias is valid",
			alias:  "xxxx",
			expErr: nil,
		},
		{
			name:   "alias is too short",
			alias:  "xxx",
			expErr: ErrInvalidAlias,
		},
		{
			name:   "alias contains not allowed characters",
			alias:  "$xxx",
			expErr: ErrInvalidAlias,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			s := New(config.Service{}, nil, nil)

			is.Equal(tc.expErr, s.ValidateCustomAlias(tc.alias))
		})
	}
}
