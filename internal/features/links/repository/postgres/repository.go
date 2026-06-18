package links_postgres_repository

import core_postgres_pool "github.com/d1mas2k3/url_shortener/internal/core/repository/postgres/pool"

// LinksPostgresRepository — postgres реализация LinksRepository
type LinksPostgresRepository struct {
	pool core_postgres_pool.Pool
}

func NewLinksPostgresRepository(
	pool core_postgres_pool.Pool,
) *LinksPostgresRepository {
	return &LinksPostgresRepository{
		pool: pool,
	}
}
