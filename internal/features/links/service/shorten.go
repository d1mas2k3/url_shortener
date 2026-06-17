// Реализует метод shorten у интерфейса LinksService. Бизнес-логика создания короткой ссылки
package links_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
)

// Возвращает короткий url (code) при заданном длинном (originalURL)
func (s *LinkService) Shorten(ctx context.Context, originalURL string) (domain.Link, error) {
	// Если originalURL пустой - возвращаем ошибку
	if len(originalURL) == 0 {
		return domain.Link{},
			fmt.Errorf(
				"shorten: url is empty: %w",
				core_errors.ErrInvalidArgument,
			)
	}

	// Если уже есть url - возвращаем его
	link, err := s.linksRepository.GetByURL(ctx, originalURL)
	if err == nil {
		return link, nil
	}

	// Если инфра-ошибка (сбой на уровне хранилища или сети)
	if !errors.Is(err, core_errors.ErrNotFound) {
		return domain.Link{}, fmt.Errorf("get by url: %w", err)
	}

	// Генерируем код для нашего URL
	for {
		code, err := generateCode()
		if err != nil {
			return domain.Link{}, fmt.Errorf("generate code: %w", err)
		}

		err = s.linksRepository.Save(
			ctx,
			domain.Link{
				Code:        code,
				OriginalURL: originalURL,
			})
		// Если такой код существует - ретрай (заново генерируем)
		if errors.Is(err, core_errors.ErrCodeExists) {
			continue
		}

		// Если на моменте генерации кода у нас уже сохранился URL, то возвращаем его и код (против гонки данных)
		if errors.Is(err, core_errors.ErrURLExists) {
			link, err := s.linksRepository.GetByURL(ctx, originalURL)
			if err != nil {
				return domain.Link{}, fmt.Errorf("get by url after race: %w", err)
			}
			return link, nil
		}

		// Проверяем, что все получилось сохранить (фун-я Save())
		if err != nil {
			return domain.Link{}, fmt.Errorf("save link: %w", err)
		}

		return domain.Link{Code: code, OriginalURL: originalURL}, nil
	}
}
