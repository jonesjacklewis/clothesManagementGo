package repository

import (
	"clothes_management/internal/domain"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type InMemoryClothingRepository struct {
	// items contains a key for userId, which contains a map of clothingItems keyed by its own id
	items map[string]map[string]domain.Clothing
	mu    sync.Mutex
}

func (r *InMemoryClothingRepository) Save(userId string, clothing domain.Clothing) (domain.Clothing, error) {

	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()
	clothing.UserId = userId
	clothing.Id = id

	_, exists := r.items[userId]

	if !exists {
		r.items[userId] = map[string]domain.Clothing{}
	}

	r.items[userId][id] = clothing

	return clothing, nil
}

func (r *InMemoryClothingRepository) GetAll(userId string) ([]domain.Clothing, error) {
	if strings.TrimSpace(userId) == "" {
		return []domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var items []domain.Clothing = []domain.Clothing{}

	userItems, exists := r.items[userId]

	if !exists {
		return items, nil
	}

	for _, item := range userItems {
		items = append(items, item)
	}

	return items, nil
}

func (r *InMemoryClothingRepository) GetById(userId, id string) (domain.Clothing, error) {
	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	userItems, exists := r.items[userId]

	if !exists {
		return domain.Clothing{}, fmt.Errorf("No item exists for id %s for user %s", id, userId)
	}

	item, exists := userItems[id]

	if !exists {
		return domain.Clothing{}, fmt.Errorf("No item exists for id %s for user %s", id, userId)
	}

	return item, nil
}

func (r *InMemoryClothingRepository) Update(userId string, clothing domain.Clothing) (domain.Clothing, error) {
	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[userId]

	if !exists {
		return domain.Clothing{}, fmt.Errorf("User ID %s not found", userId)
	}

	_, exists = r.items[userId][clothing.Id]

	if !exists {
		return domain.Clothing{}, fmt.Errorf("No item exists for id %s", clothing.Id)
	}

	r.items[userId][clothing.Id] = clothing

	return clothing, nil
}

func (r *InMemoryClothingRepository) Delete(userId, id string) error {
	if strings.TrimSpace(userId) == "" {
		return errors.New("User ID must not be empty or whitespace")
	}

	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("ID cannot be empty or whitespace")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[userId]

	if !exists {
		return fmt.Errorf("User ID %s not found", userId)
	}

	_, exists = r.items[userId][id]

	if !exists {
		return fmt.Errorf("Item with id %s does not exist", id)
	}

	delete(r.items[userId], id)

	return nil
}

func (r *InMemoryClothingRepository) Exists(userId, id string) (bool, error) {
	if strings.TrimSpace(userId) == "" {
		return false, errors.New("User ID must not be empty or whitespace")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[userId]

	if !exists {
		return false, nil
	}

	_, exists = r.items[userId][id]

	return exists, nil
}

func NewInMemoryClothingRepository() *InMemoryClothingRepository {
	return &InMemoryClothingRepository{
		items: make(map[string]map[string]domain.Clothing),
	}
}
