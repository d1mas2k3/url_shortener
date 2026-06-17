package links_memory_repository

import (
	"context"
	"fmt"

	"github.com/d1mas2k3/url-shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url-shortener/internal/core/errors"
)

func (r *LinksMemoryRepository) GetByURL(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	// Блокировка на чтение
	r.mx.RLock()
	defer r.mx.RUnlock()

	// Достаем code из byURL по originalURL (comma-ok)
	code, ok := r.byURL[originalURL]

	// Если нет code -> возвращаем пустую сущность и ошибку
	if !ok {
		return domain.Link{},
			fmt.Errorf("get by url=%q: %w",
				originalURL,
				core_errors.ErrNotFound,
			)
	}

	return domain.Link{
		Code:        code,
		OriginalURL: originalURL,
	}, nil
}
