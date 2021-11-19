package service

import (
	"context"
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
		name         string
		r            *mockLinkRepo
		url          string
		requestedURX string
		expErr       error
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
			url:          "https://xxxxxxxxxx.xxx/xxxxxx?x=xxx",
			requestedURX: "",
			expErr:       nil,
		},
		{
			name: "URL is shortened with requested URX",
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
			url:          "https://xxxxxxxxxx.xxx/xxxxxx?x=xxx",
			requestedURX: "urx",
			expErr:       nil,
		},
		{
			name:   "invalid URL",
			url:    "xxx",
			expErr: ErrInvalidURL,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			urx, err := New(config.Service{}, tc.r).Shorten(context.Background(), tc.url)

			is.Equal(tc.expErr, err)
			matched, _ := regexp.MatchString(`/[a-zA-Z0-9]{8}`, urx)
			if tc.expErr == nil && tc.requestedURX == "" && !matched {
				t.Errorf("invalid URX: %s", urx)
			}
		})
	}
}

func TestService_FindURL(t *testing.T) {
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

		url, err := New(config.Service{}, tc.r).FindURL(context.Background(), tc.urx)

		is.Equal(tc.expUrl, url)
		is.Equal(tc.expErr, err)
	}
}
