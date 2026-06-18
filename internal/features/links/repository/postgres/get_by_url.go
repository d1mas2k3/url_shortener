package links_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
	"github.com/jackc/pgx/v5"
)

func (r *LinksPostgresRepository) GetByURL(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	// Ограничивает время выполнения запроса (защита от зависаний)
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// SQL запрос, который выводит только нужные нам поля у заданного id
	query := `SELECT code, original_url FROM links WHERE original_url = $1`

	// Получаем ровно одну строку с нужными sql-данными
	row := r.pool.QueryRow(ctx, query, originalURL)
	var link domain.Link
	err := row.Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Link{},
				fmt.Errorf(
					"get by url=%q: %w",
					originalURL,
					core_errors.ErrNotFound,
				)	
		}
		return domain.Link{}, fmt.Errorf("scan get by url: %w", err)
	}
	return link, nil
}
