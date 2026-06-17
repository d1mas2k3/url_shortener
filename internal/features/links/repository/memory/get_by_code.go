package links_memory_repository

import (
	"context"
	"fmt"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
)

// GetByCode возвращает ссылку по короткому коду.
func (r *LinksMemoryRepository) GetByCode(
	ctx context.Context,
	code string,
) (domain.Link, error) {
	// Блокировка на чтение
	r.mx.RLock()
	defer r.mx.RUnlock()

	// Достаем url из byCode по code (comma-ok)
	url, ok := r.byCode[code]

	// Если нет url -> возвращаем пустую сущность и ошибку
	if !ok {
		return domain.Link{},
			fmt.Errorf("get by code=%q: %w",
				code,
				core_errors.ErrNotFound,
			)
	}

	return domain.Link{
		Code:        code,
		OriginalURL: url,
	}, nil
}
