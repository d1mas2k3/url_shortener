package links_service

import (
	"context"
	"fmt"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
)

// Resolve возвращает оригинальный URL по короткому коду.
func (s *LinkService) Resolve(ctx context.Context, code string) (domain.Link, error) {
	link, err := s.linksRepository.GetByCode(ctx, code)
	if err != nil {
		return domain.Link{}, fmt.Errorf("getting by code: %w", err)
	}

	return link, nil
}
