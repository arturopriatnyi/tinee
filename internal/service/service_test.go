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
	save        func(ctx context.Context, link Link) error
	findByURL   func(ctx context.Context, URL string) (Link, error)
	findByAlias func(ctx context.Context, alias string) (Link, error)
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

func TestNewService(t *testing.T) {
	is := is.New(t)
	cfg := config.Service{}
	r := &mockLinkRepo{}

	is.Equal(&Service{cfg: cfg, r: r}, New(cfg, r))
}

func TestService_Shorten(t *testing.T) {
	testcases := []struct {
		name   string
		r      *mockLinkRepo
		url    string
		alias  string
		expErr error
	}{
		{
			name: "URL is shortened",
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
			alias:  "",
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
			name: "find by URL unexpected error",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, errors.New("unexpected error")
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: errors.New("unexpected error"),
		},
		{
			name: "unexpected error in FindByAlias while creating task",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, errors.New("unexpected error")
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: errors.New("unexpected error"),
		},
		{
			name: "alias is taken while creating task",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					if alias == "xxxx" {
						return Link{}, nil
					}
					return Link{}, nil
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: ErrInvalidAlias,
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
			alias:  "xxx",
			expErr: ErrInvalidAlias,
		},
		{
			name: "custom alias is taken by other link",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					if alias == "xxxx" {
						return Link{ID: "x-x-x-x"}, nil
					}

					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			url:    "https://x.xx",
			alias:  "xxxx",
			expErr: ErrInvalidAlias,
		},
		{
			name: "find by alias unexpected error while setting custom alias",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByAlias: func(ctx context.Context, alias string) (Link, error) {
					if alias == "xxxx" {
						return Link{ID: "x-x-x-x"}, errors.New("unexpected error")
					}

					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
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

			urx, err := New(config.Service{}, tc.r).Shorten(context.Background(), tc.url, tc.alias)

			is.Equal(tc.expErr, err)
			matched, _ := regexp.MatchString(`/[a-zA-Z0-9]{8}`, urx)
			if tc.expErr == nil && tc.alias == "" && !matched {
				t.Errorf("invalid URX: %s", urx)
			}
		})
	}
}

func TestService_URLByAlias(t *testing.T) {
	testcases := []struct {
		name   string
		r      *mockLinkRepo
		urx    string
		expUrl string
		expErr error
	}{
		{
			name: "URL is found",
			r: &mockLinkRepo{
				findByAlias: func(ctx context.Context, URX string) (Link, error) {
					return Link{URL: "https://xxxxxxxxxx.xxx/xxxxxxxx"}, nil
				},
			},
			urx:    "xxxxxxxx",
			expUrl: "https://xxxxxxxxxx.xxx/xxxxxxxx",
			expErr: nil,
		},
	}

	for _, tc := range testcases {
		is := is.New(t)

		url, err := New(config.Service{}, tc.r).URLByAlias(context.Background(), tc.urx)

		is.Equal(tc.expUrl, url)
		is.Equal(tc.expErr, err)
	}
}
