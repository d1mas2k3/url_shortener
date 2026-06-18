package core_postgres_pool

// Вместо того чтобы каждый раз при запросе открывать новое соединение с БД — создаётся пул,
// который держит несколько соединений открытыми и переиспользует их

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Описывает что умеет pool. Репозиторий всегда работает с интерфейсом, а не с конкретной стрктурой
// Это и есть принцип dependency inversion
type Pool interface {
	// Когда ожидаешь несколько строк (например список продуктов)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// Когда ожидаешь одну строку (например один пользователь по ID)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Когда ничего не возвращается (например DELETE)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)

	// Закрывает все соединения в пуле
	Close()

	// В нашей реализации, возвращает таймаут для операций с БД
	OpTimeout() time.Duration
}

// Встраиваем *pgxpool.Pool - пул от библиотеки pgx и добавляем свой таймаут;
// Таймаут нужен для того, чтобы после n времени отменить запрос - если слишком долго, скорее всего ошибка
type ConnectionPool struct {
	*pgxpool.Pool
	opTimeout time.Duration // Параметр Timeout из структуры Config
}

// Создаёт пул соединений с PostgreSQL
func NewConnectionPool(
	ctx context.Context,
	config Config,
) (*ConnectionPool, error) {
	// 1. Собирает строку подключения
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)
	// 2. Парсит(читает строку и разбирает её на части) конфиг для pgx
	pgxconfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgxconfig: %w", err)
	}

	// 3. Создаёт пул
	pool, err := pgxpool.NewWithConfig(ctx, pgxconfig)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	// 4. Проверяет что БД реально доступна
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return &ConnectionPool{
		Pool:      pool,
		opTimeout: config.Timeout,
	}, nil
}

// Функция для возвращения timeout Нужен потому что opTimeout это приватное поле, снаружи напрямую не достать
func (p *ConnectionPool) OpTimeout() time.Duration {
	return p.opTimeout
}
