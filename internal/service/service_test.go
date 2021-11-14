package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/matryer/is"

	"urx/internal/config"
)

type mockLinkRepo struct {
	save      func(ctx context.Context, link Link) error
	findByURL func(ctx context.Context, URL string) (Link, error)
	findByURX func(ctx context.Context, URX string) (Link, error)
}

func (r *mockLinkRepo) Save(ctx context.Context, link Link) error {
	return r.save(ctx, link)
}

func (r *mockLinkRepo) FindByURL(ctx context.Context, URL string) (Link, error) {
	return r.findByURL(ctx, URL)
}

func (r *mockLinkRepo) FindByURX(ctx context.Context, URX string) (Link, error) {
	return r.findByURX(ctx, URX)
}

func TestNewService(t *testing.T) {
	is := is.New(t)
	r := &mockLinkRepo{}

	is.Equal(&Service{cfg: config.Service{Domain: "urx.io"}, r: r}, New(r))
}

func TestService_Shorten(t *testing.T) {
	testcases := []struct {
		name   string
		r      *mockLinkRepo
		url    string
		expErr error
	}{
		{
			name: "URL is shortened",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByURX: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				save: func(ctx context.Context, link Link) error {
					return nil
				},
			},
			url:    "https://xxxxxxxxxx.xxx/xxxxxx?x=xxx",
			expErr: nil,
		},
		{
			name:   "invalid URL",
			url:    "xxx",
			expErr: ErrInvalidURL,
		},
		{
			name: "URL is shortened already",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{URX: "xxxxxxxx"}, nil
				},
			},
			url:    "https://xxxxxxxxxx.xxx/xxxxxx?x=xxx",
			expErr: nil,
		},
		{
			name: "all URXs are taken",
			r: &mockLinkRepo{
				findByURL: func(ctx context.Context, url string) (Link, error) {
					return Link{}, ErrLinkNotFound
				},
				findByURX: func(ctx context.Context, urx string) (Link, error) {
					return Link{}, nil
				},
			},
			url:    "https://xxxxxxxxxx.xxx/xxxxxx?x=xxx",
			expErr: ErrGeneratingURX,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			s := New(tc.r)

			urx, err := s.Shorten(context.Background(), tc.url)

			is.Equal(tc.expErr, err)
			matched, _ := regexp.MatchString(`urx.io/[a-z0-9]{8}`, urx)
			if tc.expErr == nil && !matched {
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
				findByURX: func(ctx context.Context, URX string) (Link, error) {
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

		url, err := New(tc.r).FindURL(context.Background(), tc.urx)

		is.Equal(tc.expUrl, url)
		is.Equal(tc.expErr, err)
	}
}
