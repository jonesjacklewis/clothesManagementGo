package repository

import (
	"clothes_management/internal/domain"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type InMemoryClothingRepository struct {
	// items contains a key for userId, which contains a map of clothingItems keyed by its own id
	items               map[string]map[string]domain.Clothing
	issueOnSaveError    bool
	issueOnGetAllError  bool
	issueOnGetByIdError bool
	issueOnUpdateError  bool
	issueOnDeleteError  bool
	issueOnExistsError  bool
	mu                  sync.Mutex
}

func (r *InMemoryClothingRepository) SetIssueOnSaveError(value bool) {
	r.mu.Lock()
	r.issueOnSaveError = value
	r.mu.Unlock()
}

func (r *InMemoryClothingRepository) SetIssueOnGetAllError(value bool) {
	r.mu.Lock()
	r.issueOnGetAllError = value
	r.mu.Unlock()
}

func (r *InMemoryClothingRepository) SetIssueOnGetByIdError(value bool) {
	r.mu.Lock()
	r.issueOnGetByIdError = value
	r.mu.Unlock()
}

func (r *InMemoryClothingRepository) SetIssueOnUpdateError(value bool) {
	r.mu.Lock()
	r.issueOnUpdateError = value
	r.mu.Unlock()
}

func (r *InMemoryClothingRepository) SetIssueOnDeleteError(value bool) {
	r.mu.Lock()
	r.issueOnDeleteError = value
	r.mu.Unlock()
}

func (r *InMemoryClothingRepository) Save(userId string, clothing domain.Clothing) (domain.Clothing, error) {

	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	if r.issueOnSaveError {
		return domain.Clothing{}, errors.New("Error saving clothing item")
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

	if r.issueOnGetAllError {
		return []domain.Clothing{}, errors.New("Error getting clothing items")
	}

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

	log.Printf("r.issueOnError = %v", r.issueOnGetByIdError)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.issueOnGetByIdError {
		return domain.Clothing{}, fmt.Errorf("Unable to get clothing for ID %s", id)
	}

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

	if r.issueOnUpdateError {
		return domain.Clothing{}, errors.New("Error updating clothing item")
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

	if r.issueOnDeleteError {
		return fmt.Errorf("Unable to delete clothing for ID %s", id)
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
