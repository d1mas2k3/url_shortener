package links_transport_http

// принимает запросы, мапит DTO <-> domain и объявляет контракт сервиса LinksService

import (
	"context"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
)

type LinksHTTPHandler struct {
	linksService LinksService
}

type LinksService interface {
	Shorten(
		ctx context.Context,
		originalURL string,
	) (domain.Link, error)
	Resolve(
		ctx context.Context,
		code string,
	) (domain.Link, error)
}

func NewLinksHTTPHandler(
	linksService LinksService,
) *LinksHTTPHandler {
	return &LinksHTTPHandler{
		linksService: linksService,
	}
}
