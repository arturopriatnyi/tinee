// Package http provides HTTP handler for urx.
package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"urx/internal/service"
)

// Service is urx service interface.
type Service interface {
	Shorten(ctx context.Context, URL string) (URX string, err error)
	FindURL(ctx context.Context, URX string) (URL string, err error)
}

// Handler is HTTP handler for urx.
type Handler struct {
	r *chi.Mux
	s Service
}

// NewHandler creates and returns a new Handler instance.
func NewHandler(s Service) *Handler {
	h := &Handler{r: chi.NewRouter(), s: s}

	h.r.Get("/api/v1/shorten", h.Shorten)
	h.r.Get("/{urx}", h.Redirect)

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

// Shorten is endpoint for shortening URLs.
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")

	URX, err := h.s.Shorten(r.Context(), URL)
	if err == service.ErrInvalidURL || err == service.ErrInvalidAlias {
		h.respond(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	} else if err != nil {
		h.respond(w, http.StatusInternalServerError, nil)
	} else {
		h.respond(w, http.StatusOK, map[string]interface{}{
			"urx": URX,
		})
	}
}

// Redirect is endpoint for redirecting URXs.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	urx := chi.URLParam(r, "urx")

	url, err := h.s.FindURL(r.Context(), urx)

	if err == service.ErrLinkNotFound {
		h.respond(w, http.StatusNotFound, nil)
	} else if err != nil {
		h.respond(w, http.StatusInternalServerError, nil)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}
