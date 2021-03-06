package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"

	"tinee/internal/service"
)

type mockService struct {
	shorten     func(ctx context.Context, URL, alias string) (tineeURL string, err error)
	linkByAlias func(ctx context.Context, alias string) (l service.Link, err error)
}

func (s *mockService) Shorten(ctx context.Context, URL, alias string) (string, error) {
	return s.shorten(ctx, URL, alias)
}

func (s *mockService) LinkByAlias(ctx context.Context, alias string) (l service.Link, err error) {
	return s.linkByAlias(ctx, alias)
}

func TestHandler_Shorten(t *testing.T) {
	testcases := []struct {
		name    string
		s       Service
		body    string
		expCode int
		expBody string
	}{
		{
			name: "URL is shortened",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (tineeURL string, err error) {
					return "tinee.io/xxxxxxxx", nil
				},
			},
			body:    `{"url":"https://x.xx"}`,
			expCode: http.StatusOK,
			expBody: `{"tineeUrl":"tinee.io/xxxxxxxx"}`,
		},
		{
			name: "URL is shortened with custom alias",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (tineeURL string, err error) {
					return fmt.Sprintf("tinee.io/%s", alias), nil
				},
			},
			body:    `{"url":"https://x.xx","alias":"xxxx"}`,
			expCode: http.StatusOK,
			expBody: `{"tineeUrl":"tinee.io/xxxx"}`,
		},
		{
			name:    "empty request body",
			expCode: http.StatusBadRequest,
			expBody: `{"error":"EOF"}`,
		},
		{
			name: "invalid URL",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (tineeURL string, err error) {
					return "", service.ErrInvalidURL
				},
			},
			body:    `{"url":"x.xx"}`,
			expCode: http.StatusBadRequest,
			expBody: `{"error":"invalid URL"}`,
		},
		{
			name: "invalid alias",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (tineeURL string, err error) {
					return "", service.ErrInvalidAlias
				},
			},
			body:    `{"url":"https://x.xx","alias":"x"}`,
			expCode: http.StatusBadRequest,
			expBody: `{"error":"invalid alias"}`,
		},
		{
			name: "unexpected error",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (tineeURL string, err error) {
					return "", errors.New("unexpected error")
				},
			},
			body:    `{"url":"https://x.xx"}`,
			expCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			h := NewHandler(tc.s)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBufferString(tc.body))
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, r)

			is.Equal(tc.expCode, rr.Code)
			is.Equal(tc.expBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	testcases := []struct {
		name    string
		s       Service
		expCode int
		expURL  string
	}{
		{
			name: "tineeURL redirected to actual URL",
			s: &mockService{
				linkByAlias: func(ctx context.Context, alias string) (l service.Link, err error) {
					return service.Link{URL: "https://x.xx"}, nil
				},
			},
			expCode: http.StatusSeeOther,
			expURL:  "https://x.xx",
		},
		{
			name: "link not found",
			s: &mockService{
				linkByAlias: func(ctx context.Context, alias string) (l service.Link, err error) {
					return service.Link{}, service.ErrLinkNotFound
				},
			},
			expCode: http.StatusNotFound,
		},
		{
			name: "unexpected error",
			s: &mockService{
				linkByAlias: func(ctx context.Context, alias string) (l service.Link, err error) {
					return service.Link{}, errors.New("unexpected error")
				},
			},
			expCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			h := NewHandler(tc.s)

			r := httptest.NewRequest(http.MethodGet, "/alias", nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, r)

			is.Equal(tc.expCode, rr.Code)
			if tc.expCode == http.StatusSeeOther {
				is.Equal(tc.expURL, rr.Header().Get("Location"))
			}
		})
	}
}
