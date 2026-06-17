package links_memory_repository

// Хранит ссылки в потокобезопасных мапах (два индекса), без внешней БД.

import (
	"context"
	"fmt"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
)

// Сохраняет пару [code, url] в хранилище(LinksMemoryRepository)
func (r *LinksMemoryRepository) Save(ctx context.Context, link domain.Link) error {
	// Полный лок на запись
	r.mx.Lock()
	defer r.mx.Unlock()

	// Если link.Code уже есть в byCode -> возвращаем core_errors.ErrCodeExists
	if _, ok := r.byCode[link.Code]; ok {
		return fmt.Errorf(
			"save code=%q: %w",
			link.Code,
			core_errors.ErrCodeExists,
		)
	}

	// Если link.OriginalURL уже есть в byURL -> возвращаем core_errors.ErrURLExists
	if _, ok := r.byURL[link.OriginalURL]; ok { // Если true (есть)
		return fmt.Errorf(
			"save url=%q: %w", // %q как строки, только удобно оборачивает в ковычки
			link.OriginalURL,
			core_errors.ErrURLExists,
		)
	}

	// Записась в обе мапы
	r.byURL[link.OriginalURL] = link.Code
	r.byCode[link.Code] = link.OriginalURL

	return nil
}
