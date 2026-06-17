package links_transport_http

// принимает запросы, мапит DTO <-> domain и объявляет контракт сервиса LinksService

import (
	"context"
	"net/http"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_http_server "github.com/d1mas2k3/url_shortener/internal/core/transport/http/server"
)

type LinksHTTPHandler struct {
	linksService LinksService
	baseURL      string
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

// Возвращает список HTTP-маршрутов (эндпоинтов), которые поддерживает LinksHTTPHandler
func (h *LinksHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/links",
			Handler: h.Shorten,
		},
		{
			Method:  http.MethodGet,
			Path:    "/links/{code}",
			Handler: h.Resolve,
		},
	}
}

func NewLinksHTTPHandler(
	linksService LinksService,
	baseURL string,
) *LinksHTTPHandler {
	return &LinksHTTPHandler{
		linksService: linksService,
		baseURL:      baseURL,
	}
}
