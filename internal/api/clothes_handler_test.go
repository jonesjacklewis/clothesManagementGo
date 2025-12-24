package api

import (
	"bytes"
	"clothes_management/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type DummyClothingRepo struct {
	ExpectedSaveItem *domain.Clothing
	SaveError        error
	NextId           string
	AllItems         []domain.Clothing
	GetAllError      error
}

func (d *DummyClothingRepo) Save(clothing domain.Clothing) (domain.Clothing, error) {
	if d.SaveError != nil {
		return domain.Clothing{}, d.SaveError
	}

	d.ExpectedSaveItem = &clothing

	if d.NextId != "" {
		clothing.Id = d.NextId
	} else {
		clothing.Id = "dummy-id-123"
	}

	return clothing, nil
}

func (d *DummyClothingRepo) GetAll() ([]domain.Clothing, error) {
	if d.GetAllError != nil {
		return []domain.Clothing{}, d.GetAllError
	}

	var items []domain.Clothing = []domain.Clothing{}

	for _, item := range d.AllItems {
		items = append(items, item)
	}

	return items, nil
}

func TestClothes(t *testing.T) {
	t.Run("Given request doesn't use POST or GET, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodDelete, "/clothes", nil)
		r.Header.Set("Content-Type", "application/json")

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected a %d error, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unauthorised method %s. The supported methods are POST or GET.", http.MethodDelete)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, with no or empty body, should return an error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPost, "/clothes", http.NoBody)
		r.Header.Set("Content-Type", "application/json")

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must not be empty or missing"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expecteed %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data isn't json, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPost, "/clothes", strings.NewReader("data"))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must be JSON"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has no pricePence, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"clothingType": "Jumper",
			"description":  "Red Loosefit Jumper",
			"brand":        "A&B",
			"store":        "Totally Real Store",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'pricePence'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has no clothingType, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":  2000,
			"description": "Red Loosefit Jumper",
			"brand":       "A&B",
			"store":       "Totally Real Store",
			"size":        "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'clothingType'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has no description, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"brand":        "A&B",
			"store":        "Totally Real Store",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'description'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has no brand, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"store":        "Totally Real Store",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'brand'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has no store, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'store'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has no size, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"store":        "Totlly Real Store",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'size'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has extra field, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"store":        "Totlly Real Store",
			"size":         "Medium",
			"fakeField":    "fakeValue",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, superfluous fields"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data is structurally valid, but breaks validation, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   -2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"store":        "Totlly Real Store",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, breaks validation rule:"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, with valid data but issue on save, should have error", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"store":        "Totlly Real Store",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))

		dummyRepo := &DummyClothingRepo{
			SaveError: fmt.Errorf("A dummy error"),
		}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected %d got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		expectedMessage := "Error saving clothing item"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}

	})

	t.Run("Given POST request, with valid data and no issue on save, should have success", func(t *testing.T) {
		w := httptest.NewRecorder()

		var jsonMap map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"store":        "Totlly Real Store",
			"size":         "Medium",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/clothes", bytes.NewReader(body))
		expectedID := "generated-test-id-1"
		dummyRepo := &DummyClothingRepo{
			NextId: expectedID,
		}
		apiHandler := &API{
			Repo: dummyRepo,
		}
		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected %d got %d", http.StatusCreated, resp.StatusCode)
		}

		if dummyRepo.ExpectedSaveItem == nil {
			t.Error("Expected Save to be called, but ExpectedSaveItem is nil")
		} else {
			if dummyRepo.ExpectedSaveItem.ClothingType != jsonMap["clothingType"] {
				t.Errorf("Expected saved ClothingType %v, got %v", jsonMap["clothingType"], dummyRepo.ExpectedSaveItem.ClothingType)
			}
		}

		var responseBody map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &responseBody)

		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		if responseBody["success"] != true {
			t.Errorf("Expected success: true, got %v", responseBody["success"])
		}

		data, ok := responseBody["data"].(map[string]any)
		if !ok {
			t.Fatalf("Expected 'data' field in response, got %v", responseBody["data"])
		}

		if data["id"] != expectedID {
			t.Errorf("Expected returned ID %s, got %v", expectedID, data["id"])
		}

		if data["clothingType"] != jsonMap["clothingType"] {
			t.Errorf("Expected returned clothingType %s, got %v", jsonMap["clothingType"], data["clothingType"])
		}

	})

	t.Run("Given GET request, with issues with retrieval, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)

		dummyRepo := &DummyClothingRepo{
			GetAllError: errors.New("Some Type of Error"),
		}
		apiHandler := &API{
			Repo: dummyRepo,
		}

		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expeted %d got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		expected := "Error getting clothing items"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}

	})

	t.Run("Given GET request, with no issues with retrieval, should return success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)

		expectedItems := []domain.Clothing{
			{Id: "id-1", ClothingType: "Shirt", Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M"},
			{Id: "id-2", ClothingType: "Trousers", Description: "Black Jeans", Brand: "Y", Store: "Store", Price: 2500, Size: "L"},
		}
		dummyRepo := &DummyClothingRepo{
			AllItems: expectedItems,
		}
		apiHandler := &API{
			Repo: dummyRepo,
		}

		apiHandler.Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expeted %d got %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)

		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		if responseBody["success"] != true {
			t.Errorf("Expected success: true, got %v", responseBody["success"])
		}

		data, ok := responseBody["data"].([]any)
		if !ok {
			t.Fatalf("Expected 'data' field in response, got %v", responseBody["data"])
		}

		if len(data) != len(expectedItems) {
			t.Errorf("Expected %d items, got %d", len(expectedItems), len(data))
		}
	})

}
