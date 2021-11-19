package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"

	"urx/internal/service"
)

type mockService struct {
	shorten func(ctx context.Context, URL string) (URX string, err error)
	findURL func(ctx context.Context, URX string) (URL string, err error)
}

func (s *mockService) Shorten(ctx context.Context, URL string) (URX string, err error) {
	return s.shorten(ctx, URL)
}

func (s *mockService) FindURL(ctx context.Context, URX string) (URL string, err error) {
	return s.findURL(ctx, URX)
}

func TestHandler_Shorten(t *testing.T) {
	testcases := []struct {
		name         string
		s            Service
		url          string
		requestedURX string
		expCode      int
		expBody      string
	}{
		{
			name: "URL is shortened",
			s: &mockService{
				shorten: func(ctx context.Context, URL string) (URX string, err error) {
					return "urx.io/xxxxxxxx", nil
				},
			},
			expCode: http.StatusOK,
			expBody: `{"urx":"urx.io/xxxxxxxx"}`,
		},
		{
			name: "invalid URL",
			s: &mockService{
				shorten: func(ctx context.Context, URL string) (URX string, err error) {
					return "", service.ErrInvalidURL
				},
			},
			expCode: http.StatusBadRequest,
			expBody: `{"error":"invalid URL"}`,
		},
		{
			name: "unexpected error",
			s: &mockService{
				shorten: func(ctx context.Context, URL string) (URX string, err error) {
					return "", errors.New("unexpected error")
				},
			},
			expCode: http.StatusInternalServerError,
			expBody: ``,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			h := NewHandler(tc.s)

			r := httptest.NewRequest(http.MethodGet, "/api/v1/shorten", nil)
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
	}{
		{
			name: "URX redirected",
			s: &mockService{
				findURL: func(ctx context.Context, URX string) (URL string, err error) {
					return "https://xxxxxxxxxx.xxx/xxx?x=x", nil
				},
			},
			expCode: http.StatusSeeOther,
		},
		{
			name: "URX not found",
			s: &mockService{
				findURL: func(ctx context.Context, URX string) (URL string, err error) {
					return "", service.ErrLinkNotFound
				},
			},
			expCode: http.StatusNotFound,
		},
		{
			name: "unexpected error",
			s: &mockService{
				findURL: func(ctx context.Context, URX string) (URL string, err error) {
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

			r := httptest.NewRequest(http.MethodGet, "/urx", nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, r)

			is.Equal(tc.expCode, rr.Code)
		})
	}
}
