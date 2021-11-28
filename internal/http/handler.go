// Package http provides HTTP handler for urx.
package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"urx/internal/service"
)

// Service is urx service interface.
type Service interface {
	Shorten(ctx context.Context, URL, alias string) (URX string, err error)
	URLByAlias(ctx context.Context, alias string) (URL string, err error)
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
	h.r.Get("/{alias}", h.Redirect)

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
	alias := r.URL.Query().Get("alias")

	URX, err := h.s.Shorten(r.Context(), URL, alias)
	if err == service.ErrInvalidURL || err == service.ErrInvalidAlias {
		h.respond(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	} else if err != nil {
		zap.L().Error(err.Error())
		h.respond(w, http.StatusInternalServerError, nil)
	} else {
		h.respond(w, http.StatusOK, map[string]interface{}{
			"urx": URX,
		})
	}
}

// Redirect is endpoint for redirecting URXs.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")

	URL, err := h.s.URLByAlias(r.Context(), alias)
	if err == service.ErrLinkNotFound {
		h.respond(w, http.StatusNotFound, nil)
	} else if err != nil {
		zap.L().Error(err.Error())
		h.respond(w, http.StatusInternalServerError, nil)
	} else {
		http.Redirect(w, r, URL, http.StatusSeeOther)
	}
}
