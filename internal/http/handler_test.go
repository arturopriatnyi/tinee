package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"

	"urx/internal/service"
)

type mockService struct {
	shorten    func(ctx context.Context, URL, alias string) (URX string, err error)
	urlByAlias func(ctx context.Context, alias string) (URL string, err error)
}

func (s *mockService) Shorten(ctx context.Context, URL, alias string) (URX string, err error) {
	return s.shorten(ctx, URL, alias)
}

func (s *mockService) URLByAlias(ctx context.Context, alias string) (URL string, err error) {
	return s.urlByAlias(ctx, alias)
}

func TestHandler_Shorten(t *testing.T) {
	testcases := []struct {
		name        string
		s           Service
		queryParams string
		expCode     int
		expBody     string
	}{
		{
			name: "URL is shortened",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (URX string, err error) {
					return "urx.io/xxxxxxxx", nil
				},
			},
			queryParams: "?url=https://x.xx",
			expCode:     http.StatusOK,
			expBody:     `{"urx":"urx.io/xxxxxxxx"}`,
		},
		{
			name: "URL is shortened with custom alias",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (URX string, err error) {
					return fmt.Sprintf("urx.io/%s", alias), nil
				},
			},
			queryParams: "?url=https://x.xx&alias=xxxx",
			expCode:     http.StatusOK,
			expBody:     `{"urx":"urx.io/xxxx"}`,
		},
		{
			name: "invalid URL",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (URX string, err error) {
					return "", service.ErrInvalidURL
				},
			},
			expCode: http.StatusBadRequest,
			expBody: `{"error":"invalid URL"}`,
		},
		{
			name: "invalid alias",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (URX string, err error) {
					return "", service.ErrInvalidAlias
				},
			},
			expCode: http.StatusBadRequest,
			expBody: `{"error":"invalid alias"}`,
		},
		{
			name: "unexpected error",
			s: &mockService{
				shorten: func(ctx context.Context, URL, alias string) (URX string, err error) {
					return "", errors.New("unexpected error")
				},
			},
			expCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			h := NewHandler(tc.s)

			r := httptest.NewRequest(http.MethodGet, "/api/v1/shorten"+tc.queryParams, nil)
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
			name: "URX redirected to actual URL",
			s: &mockService{
				urlByAlias: func(ctx context.Context, alias string) (URL string, err error) {
					return "https://x.xx", nil
				},
			},
			expCode: http.StatusSeeOther,
			expURL:  "https://x.xx",
		},
		{
			name: "URX not found",
			s: &mockService{
				urlByAlias: func(ctx context.Context, alias string) (URL string, err error) {
					return "", service.ErrLinkNotFound
				},
			},
			expCode: http.StatusNotFound,
		},
		{
			name: "unexpected error",
			s: &mockService{
				urlByAlias: func(ctx context.Context, alias string) (URL string, err error) {
					return "", errors.New("unexpected error")
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
