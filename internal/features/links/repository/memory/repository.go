package links_memory_repository

import "sync"

// Хранилище с двумя мапами (чтобы быстро искать в обе стороны) и Мьютекс для защищенного доступа
type LinksMemoryRepository struct {
	mx     sync.RWMutex
	byCode map[string]string
	byURL  map[string]string
}

func NewLinksMemoryRepository() *LinksMemoryRepository {
	return &LinksMemoryRepository{
		byCode: make(map[string]string),
		byURL:  make(map[string]string),
	}
}
