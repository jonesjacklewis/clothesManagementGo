package repository

import (
	"clothes_management/internal/domain"
	"fmt"
	"strings"
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

func (r *InMemoryClothingRepository) GetById(id string) (domain.Clothing, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.items[id]

	if !exists {
		return domain.Clothing{}, fmt.Errorf("No item exists for id %s", id)
	}

	return item, nil
}

func (r *InMemoryClothingRepository) Update(clothing domain.Clothing) (domain.Clothing, error) {

	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[clothing.Id]

	if !exists {
		return domain.Clothing{}, fmt.Errorf("No item exists for id %s", clothing.Id)
	}

	r.items[clothing.Id] = clothing

	return clothing, nil
}

func (r *InMemoryClothingRepository) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("ID cannot be empty or whitespace")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[id]

	if !exists {
		return fmt.Errorf("Item with id %s does not exist", id)
	}

	delete(r.items, id)

	return nil
}

func (r *InMemoryClothingRepository) Exists(id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[id]

	return exists, nil
}

func NewInMemoryClothingRepository() *InMemoryClothingRepository {
	return &InMemoryClothingRepository{
		items: make(map[string]domain.Clothing),
	}
}
