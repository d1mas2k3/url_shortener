package links_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *LinksPostgresRepository) Save(
	ctx context.Context, 
	link domain.Link,
	) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `INSERT INTO links (code, original_url) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, link.Code, link.OriginalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "links_code_key" {
				return fmt.Errorf("save code=%q: %w", link.Code, core_errors.ErrCodeExists)
			}
			return fmt.Errorf("save url=%q: %w", link.OriginalURL, core_errors.ErrURLExists)
		}
		return fmt.Errorf("exec insert link: %w", err)
	}

	return nil
}
