package repository

import (
	"clothes_management/internal/domain"
	"fmt"
	"sync"
	"testing"
)

func TestNewInMemoryClothingRepository(t *testing.T) {
	t.Run("Given the constructor is called, should create a new instance", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}
	})
}

func TestInMemorySave(t *testing.T) {
	t.Run("Given invalid clothing, should return error", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		clothingItem := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        -2000,
		}

		item, err := repo.Save(clothingItem)

		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if item.Id != "" {
			t.Errorf("Expected item not to have id, got %s", item.Id)
		}

		if len(repo.items) > 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

	})

	t.Run("Given valid clothing, should not error and should have an ID", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		clothingItem := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		item, err := repo.Save(clothingItem)

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if item.Id == "" {
			t.Error("Expected item to be populated")
		}

		if len(repo.items) == 0 {
			t.Error("Expected repo.items.length > 0")
		}

		if item.Description != clothingItem.Description {
			t.Errorf("Expected %s got %s", clothingItem.Description, item.Description)
		}

	})
}

func TestInMemoryGetAll(t *testing.T) {
	t.Run("When there are no clothing items, GetAll should return an empty list and no errors", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if len(items) > 0 {
			t.Errorf("Expected no items, got %d", len(items))
		}
	})

	t.Run("When there is 1 clothing item, GetAll should return an list with len = 1 and no errors", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		repo.mu.Lock()

		repo.items["dummy-id-123"] = domain.Clothing{
			Id:           "dummy-id-123",
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		repo.mu.Unlock()

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if len(items) != 1 {
			t.Fatalf("Expected 1 item, got %d", len(items))
		}
	})

	t.Run("When there is more than 1 clothing item, GetAll should return an list with len > 1 and no errors", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		repo.mu.Lock()

		repo.items["dummy-id-123"] = domain.Clothing{
			Id:           "dummy-id-123",
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		repo.items["dummy-id-456"] = domain.Clothing{
			Id:           "dummy-id-456",
			ClothingType: "Troussers",
			Description:  "This Troussers",
			Store:        "This Store",
			Size:         "M",
			Brand:        "XYZ",
			Price:        1500,
		}

		repo.mu.Unlock()

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if len(items) < 2 {
			t.Errorf("Expected two items, got %d", len(items))
		}
	})
}

func TestSaveAndGetAll(t *testing.T) {
	t.Run("When items are saved, GetAll should show that a new item has been added", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error (Get All 1), got %v", err)
		}

		if len(items) > 0 {
			t.Errorf("Expected no items, got %d", len(items))
		}

		clothingItem := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		_, err = repo.Save(clothingItem)

		if err != nil {
			t.Errorf("Expected not error (Save 1), got %v", err)
		}

		items, err = repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error (Get All 2), got %v", err)
		}

		if len(items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(items))
		}

		clothingItem = domain.Clothing{
			ClothingType: "Shorts",
			Description:  "This Shorts",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        1500,
		}

		_, err = repo.Save(clothingItem)

		if err != nil {
			t.Errorf("Expected not error (Save 2), got %v", err)
		}

		items, err = repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error (Get All 3), got %v", err)
		}

		if len(items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(items))
		}

	})
}

func TestInMemoryConcurrentSaves(t *testing.T) {
	t.Run("Given multiple goroutines concurrently save items, all items should be saved correctly", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		numConcurrentSaves := 100
		var wg sync.WaitGroup

		for i := 0; i < numConcurrentSaves; i++ {
			wg.Add(1)

			go func(index int) {
				defer wg.Done()

				clothingItem := domain.Clothing{
					ClothingType: fmt.Sprintf("Jumper %d", index),
					Description:  fmt.Sprintf("Concurrent Jumper %d", index),
					Brand:        "ConcurrentBrand",
					Store:        "ConcurrentStore",
					Size:         "M",
					Price:        domain.Pence(1000 + index),
				}

				_, err := repo.Save(clothingItem)

				if err != nil {
					t.Errorf("Goroutine %d: Failed to save item: %v", index, err)
				}

			}(i)

		}

		wg.Wait()

		allSavedItems, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed to retrieve all items after concurrent saves: %v", err)
		}

		if len(allSavedItems) != numConcurrentSaves {
			t.Errorf("Expected %d items after concurrent saves, got %d", numConcurrentSaves, len(allSavedItems))
		}
	})
}

func TestInMemoryConcurrentGetAll(t *testing.T) {
	t.Run("Given multiple goroutines concurrently retrieve items, all retrievals should be consistent and error-free", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		numInitialItems := 10
		for i := 0; i < numInitialItems; i++ {
			clothingItem := domain.Clothing{
				ClothingType: fmt.Sprintf("Initial Jumper %d", i),
				Description:  fmt.Sprintf("Initial Description %d", i),
				Brand:        "InitialBrand",
				Store:        "InitialStore",
				Size:         "S",
				Price:        domain.Pence(500 + i),
			}
			_, err := repo.Save(clothingItem)
			if err != nil {
				t.Fatalf("Failed to setup initial item for concurrent GetAll test: %v", err)
			}
		}

		numConcurrentReads := 500
		var wg sync.WaitGroup
		for i := 0; i < numConcurrentReads; i++ {
			wg.Add(1)
			go func(readIndex int) {
				defer wg.Done()

				items, err := repo.GetAll()
				if err != nil {
					t.Errorf("Goroutine %d: Failed to retrieve items: %v", readIndex, err)
					return
				}

				if len(items) != numInitialItems {
					t.Errorf("Goroutine %d: Expected %d items, got %d", readIndex, numInitialItems, len(items))
				}
			}(i)
		}

		wg.Wait()

		finalItems, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed final GetAll check: %v", err)
		}
		if len(finalItems) != numInitialItems {
			t.Errorf("Final check: Expected %d items, got %d", numInitialItems, len(finalItems))
		}
	})
}
