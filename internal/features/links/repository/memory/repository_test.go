package links_memory_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/d1mas2k3/url_shortener/internal/core/domain"
	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
)

// Создаем экземпляр, и сохраняем его; не получилось - ошибка
func TestSave_Success(t *testing.T) {
	repo := NewLinksMemoryRepository()
	link := domain.Link{Code: "abc1234567", OriginalURL: "https://example.com"}

	err := repo.Save(context.Background(), link)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// Сохраняем ссылку, потом создаем другой объект с той же ссылкой, и ожидаем ошибку;
// Если ошибка есть (смогли поймать ошибку) - все ок; иначе - t.Fatalf
func TestSave_DuplicateURL(t *testing.T) {
	repo := NewLinksMemoryRepository()
	link := domain.Link{Code: "abc1234567", OriginalURL: "https://example.com"}

	_ = repo.Save(context.Background(), link)

	link2 := domain.Link{Code: "xyz9999999", OriginalURL: "https://example.com"}
	err := repo.Save(context.Background(), link2)

	if !errors.Is(err, core_errors.ErrURLExists) {
		t.Fatalf("expected ErrURLExists, got: %v", err)
	}
}

// Тоже самое как предыдущая функция, но теперь проверка на коллизию кода
func TestSave_DuplicateCode(t *testing.T) {
	repo := NewLinksMemoryRepository()
	link := domain.Link{Code: "abc1234567", OriginalURL: "https://example.com"}

	_ = repo.Save(context.Background(), link)

	link2 := domain.Link{Code: "abc1234567", OriginalURL: "https://other.com"}
	err := repo.Save(context.Background(), link2)

	if !errors.Is(err, core_errors.ErrCodeExists) {
		t.Fatalf("expected ErrCodeExists, got: %v", err)
	}
}

// Сохраняем репозиторий (LinksMemoryRepository); Достаем code из репозитория и проверяем,
// что он совпадает с тем, который мы положили в него
func TestGetByCode_Success(t *testing.T) {
	repo := NewLinksMemoryRepository()
	link := domain.Link{Code: "abc1234567", OriginalURL: "https://example.com"}
	_ = repo.Save(context.Background(), link)

	got, err := repo.GetByCode(context.Background(), "abc1234567")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got.Code != link.Code || got.OriginalURL != link.OriginalURL {
		t.Fatalf("expected %+v, got %+v", link, got)
	}
}

// Проверяем, что при запросе кода (code), которого нет в хранилище будет ошибка
func TestGetByCode_NotFound(t *testing.T) {
	repo := NewLinksMemoryRepository()

	_, err := repo.GetByCode(context.Background(), "nonexistent")

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

// Сохраняем репозиторий (LinksMemoryRepository); Достаем URL из репозитория и проверяем,
// что он совпадает с тем, который мы положили в него
func TestGetByURL_Success(t *testing.T) {
	repo := NewLinksMemoryRepository()
	link := domain.Link{Code: "abc1234567", OriginalURL: "https://example.com"}
	_ = repo.Save(context.Background(), link)

	got, err := repo.GetByURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got.Code != link.Code || got.OriginalURL != link.OriginalURL {
		t.Fatalf("expected %+v, got %+v", link, got)
	}
}

// Проверяем, что при запросе url, которого нет в хранилище, будет ошибка
func TestGetByURL_NotFound(t *testing.T) {
	repo := NewLinksMemoryRepository()

	_, err := repo.GetByURL(context.Background(), "https://nonexistent.com")

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
