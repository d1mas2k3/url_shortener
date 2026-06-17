package links_service

import (
	"context"
	"errors"
	"testing"

	links_memory_repository "github.com/d1mas2k3/url_shortener/internal/features/links/repository/memory"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
)

// Сохраняем новый url, проверяем что вернулся код длиной 10.
func TestShorten_Success(t *testing.T) {
	repo := links_memory_repository.NewLinksMemoryRepository()
	svc := NewLinksService(repo)

	got, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(got.Code) != codeLength {
		t.Fatalf("expected code length %d, got: %d", codeLength, len(got.Code))
	}
}

// Пустой URL -> ErrInvalidArgument
func TestShorten_EmptyURL(t *testing.T) {
	repo := links_memory_repository.NewLinksMemoryRepository()
	svc := NewLinksService(repo)

	_, err := svc.Shorten(context.Background(), "")
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

// Один и тот же URL дважды -> оба раза возвращается один и тот же код (дедупликация)
// Другими словами, проверяем что мы не по кд создаем новый код, а берем готовый если есть
func TestShorten_ExistingURL(t *testing.T) {
	repo := links_memory_repository.NewLinksMemoryRepository()
	svc := NewLinksService(repo)

	first, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	second, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if first.Code != second.Code {
		t.Fatalf("expected same code, got %s and %s", first.Code, second.Code)
	}
}

// Из url в code, потом из code в url; потом проверяем, что первоначальный url совпадает с полученным
func TestResolve_Success(t *testing.T) {
	repo := links_memory_repository.NewLinksMemoryRepository()
	svc := NewLinksService(repo)

	saved, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("expected no error on Shorten, got: %v", err)
	}

	got, err := svc.Resolve(context.Background(), saved.Code)
	if err != nil {
		t.Fatalf("expected no error on Resolve, got: %v", err)
	}
	if got.OriginalURL != "https://example.com" {
		t.Fatalf("expected OriginalURL=https://example.com, got: %s", got.OriginalURL)
	}
}

// Resolve() по несуществующему коду - ErrNotFound
func TestResolve_NotFound(t *testing.T) {
	repo := links_memory_repository.NewLinksMemoryRepository()
	svc := NewLinksService(repo)

	_, err := svc.Resolve(context.Background(), "nonexistent")
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
