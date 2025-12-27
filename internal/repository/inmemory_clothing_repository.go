package repository

import (
	"clothes_management/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type InMemoryClothingRepository struct {
	items map[string]domain.Clothing
	mu    sync.Mutex
}

func (r *InMemoryClothingRepository) Save(clothing domain.Clothing) (domain.Clothing, error) {
	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()

	clothing.Id = id

	r.items[id] = clothing

	return clothing, nil
}

func (r *InMemoryClothingRepository) GetAll() ([]domain.Clothing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var items []domain.Clothing = []domain.Clothing{}

	for _, item := range r.items {
		items = append(items, item)
	}

	return items, nil
}

func NewInMemoryClothingRepository() *InMemoryClothingRepository {
	return &InMemoryClothingRepository{
		items: make(map[string]domain.Clothing),
	}
}
