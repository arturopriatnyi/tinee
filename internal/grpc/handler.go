// Package grpc provides gRPC Handler.
package grpc

import (
	"context"

	"urx/pkg/pb"
)

// Service is urx service interface.
type Service interface {
	Shorten(ctx context.Context, URL, alias string) (URX string, err error)
	URLByAlias(ctx context.Context, alias string) (URL string, err error)
}

// Handler is gRPC handler.
type Handler struct {
	s Service
}

// NewHandler creates and returns a new Handler instance.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Shorten shortens URL.
func (h *Handler) Shorten(ctx context.Context, r *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	URX, err := h.s.Shorten(ctx, r.GetUrl(), r.GetAlias())

	return &pb.ShortenResponse{Urx: URX}, err
}

// UrlByAlias returns URL that corresponds to alias in request.
func (h *Handler) UrlByAlias(ctx context.Context, r *pb.UrlByAliasRequest) (*pb.UrlByAliasResponse, error) {
	URL, err := h.s.URLByAlias(ctx, r.GetAlias())

	return &pb.UrlByAliasResponse{Url: URL}, err
}
