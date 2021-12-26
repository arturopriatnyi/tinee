// Package http provides HTTP handler for tinee.
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"tinee/internal/service"
)

// Service is tinee service interface.
type Service interface {
	Shorten(ctx context.Context, URL, alias string) (tineeURL string, err error)
	LinkByAlias(ctx context.Context, alias string) (l service.Link, err error)
}

// Handler is HTTP handler for tinee.
type Handler struct {
	r *chi.Mux
	s Service
}

// NewHandler creates and returns a new Handler instance.
func NewHandler(s Service) *Handler {
	h := &Handler{r: chi.NewRouter(), s: s}

	h.r.Post("/api/v1/shorten", LogResponseTime(h.Shorten))
	h.r.Get("/{alias}", LogResponseTime(h.Redirect))

	return h
}

// ServeHTTP implements standard http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

// respond responds to request.
func (h *Handler) respond(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// ShortenInput is request DTO for shortening endpoint.
type ShortenInput struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

// ShortenOutput is response DTO for shortening endpoint.
type ShortenOutput struct {
	TineeURL string `json:"tineeUrl"`
}

// Shorten is endpoint for shortening URLs.
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var i ShortenInput
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		h.respond(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	tineeURL, err := h.s.Shorten(r.Context(), i.URL, i.Alias)
	if err == service.ErrInvalidURL || err == service.ErrInvalidAlias {
		h.respond(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	} else if err != nil {
		zap.L().Error(err.Error())
		h.respond(w, http.StatusInternalServerError, nil)
	} else {
		h.respond(w, http.StatusOK, ShortenOutput{TineeURL: tineeURL})
	}
}

// Redirect is endpoint for redirecting shortened URLs.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")

	l, err := h.s.LinkByAlias(r.Context(), alias)
	if err == service.ErrLinkNotFound {
		h.respond(w, http.StatusNotFound, nil)
	} else if err != nil {
		zap.L().Error(err.Error())
		h.respond(w, http.StatusInternalServerError, nil)
	} else {
		http.Redirect(w, r, l.URL, http.StatusSeeOther)
	}
}

// LogResponseTime is middleware for logging request execution time.
func LogResponseTime(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		zap.S().Infof("request: %s, took: %v", r.URL, time.Since(start))
	}
}
