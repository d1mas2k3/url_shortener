package links_service

// Содержит бизнес-логику работы со ссылками

import (
	"context"

	"github.com/d1mas2k3/url-shortener/internal/core/domain"
)

type LinkService struct {
	linksRepository LinksRepository
}

type LinksRepository interface {
	Save(
		ctx context.Context,
		link domain.Link,
	) error
	GetByCode( // То есть по 10 символам
		ctx context.Context,
		code string,
	) (domain.Link, error)
	GetByURL( // По всему URL
		ctx context.Context,
		originalURL string,
	) (domain.Link, error)
}

func NewLinksService(
	linksRepository LinksRepository,
) *LinkService {
	return &LinkService{
		linksRepository: linksRepository,
	}
}
