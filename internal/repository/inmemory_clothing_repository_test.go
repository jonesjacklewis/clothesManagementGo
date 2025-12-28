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

func TestInMemoryGetById(t *testing.T) {
	t.Run("When the item doesn't exist GetById should return an error", func(t *testing.T) {
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

		_, err := repo.GetById("dummy-id-123")

		if err == nil {
			t.Errorf("Expected an error")
		}

		expectedMessage := "No item exists for id dummy-id-123"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When the item exists, GetById should return the item", func(t *testing.T) {
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

		dummyId := "dummy-id-123"

		repo.items[dummyId] = domain.Clothing{
			Id:           dummyId,
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		repo.mu.Unlock()

		item, err := repo.GetById(dummyId)

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if item.Id != dummyId {
			t.Errorf("Expected item.Id = %s, got %s", dummyId, item.Id)
		}
	})
}

func TestInMemoryUpdate(t *testing.T) {
	t.Run("When the item isn't valid should return an error", func(t *testing.T) {
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

		dummyId := "dummy-id-123"

		item := domain.Clothing{
			Id:           dummyId,
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        -2000,
		}

		_, err := repo.Update(item)

		if err == nil {
			t.Errorf("Expected an error")
		}

		expectedMessage := "Clothing Price must be greater than or equal to 0"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When the item doesn't exist should return an error", func(t *testing.T) {
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

		dummyId := "dummy-id-123"

		item := domain.Clothing{
			Id:           dummyId,
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		_, err := repo.Update(item)

		if err == nil {
			t.Errorf("Expected an error")
		}

		expectedMessage := "No item exists for id dummy-id-123"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When the item exists, Update should update the item", func(t *testing.T) {
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
		originalPrice := domain.Pence(2000)
		item := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        originalPrice,
		}

		item, err := repo.Save(item)

		itemId := item.Id

		updatedItem := item
		newPrice := domain.Pence(1500)
		updatedItem.Price = newPrice

		updatedItem, err = repo.Update(updatedItem)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if updatedItem.Id != itemId {
			t.Errorf("Expected the updatedItem.Id = %s, got %s", itemId, updatedItem.Id)
		}

		if updatedItem.Price == originalPrice {
			t.Errorf("Expected updatedItem.Price = %d got %d", newPrice, originalPrice)
		}

	})
}

func TestInMemoryDelete(t *testing.T) {

	t.Run("When ID is empty, should return an error", func(t *testing.T) {
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

		dummyId := ""

		err := repo.Delete(dummyId)

		if err == nil {
			t.Fatalf("Expected an error")
		}

		expectedMessage := "ID cannot be empty or whitespace"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When ID does not exist, should return an error", func(t *testing.T) {
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

		dummyId := "dummy-id-123"

		err := repo.Delete(dummyId)

		if err == nil {
			t.Fatalf("Expected an error")
		}

		expectedMessage := fmt.Sprintf("Item with id %s does not exist", dummyId)

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When ID exists, should delete successfully", func(t *testing.T) {
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

		item := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		item, err := repo.Save(item)

		if err != nil {
			t.Fatalf("Expected no error %v", err)
		}

		items, err := repo.GetAll()

		if err != nil {
			t.Fatalf("Expected no error %v", err)
		}

		itemCount := len(items)

		err = repo.Delete(item.Id)

		if err != nil {
			t.Fatalf("Expected no error %v", err)
		}

		items, err = repo.GetAll()

		if err != nil {
			t.Fatalf("Expected no error %v", err)
		}

		if len(items) != itemCount-1 {
			t.Errorf("Expected %d got %d", itemCount-1, len(items))
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
