package links_transport_http

import "github.com/d1mas2k3/url_shortener/internal/core/domain"

// ShortenRequest - тело POST-запроса
type ShortenRequest struct {
	URL string `json:"url" validate:"required"`
}

// ShortenResponse - тело ответа на POST
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// ResolveResponse - тело ответа на GET
type ResolveResponse struct {
	OriginalURL string `json:"original_url"`
}

// Склеивает ссылку с кодом вместо url
func linkDTOFromDomain(link domain.Link, baseURL string) ShortenResponse {
	// например "http://localhost:8080/abc1234567"
	return ShortenResponse{
		ShortURL: baseURL + "/" + link.Code,
	}
}
