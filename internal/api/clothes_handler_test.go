package api

import (
	"bytes"
	"clothes_management/internal/domain"
	"clothes_management/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestMissingMandatoryClothingField(t *testing.T) {
	t.Run("Given no pricePence field, should return true and appropriate message", func(t *testing.T) {
		var req map[string]any = map[string]any{}

		missing, message := MissingMandatoryClothingField(req)

		if !missing {
			t.Errorf("Expected missing to be true")
		}

		expectedMessage := "Invalid request body, missing 'pricePence'"

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})

	t.Run("Given no clothingType field, should return true and appropriate message", func(t *testing.T) {
		var req map[string]any = map[string]any{
			"pricePence": 2000,
		}

		missing, message := MissingMandatoryClothingField(req)

		if !missing {
			t.Errorf("Expected missing to be true")
		}

		expectedMessage := "Invalid request body, missing 'clothingType'"

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})

	t.Run("Given no description field, should return true and appropriate message", func(t *testing.T) {
		var req map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
		}

		missing, message := MissingMandatoryClothingField(req)

		if !missing {
			t.Errorf("Expected missing to be true")
		}

		expectedMessage := "Invalid request body, missing 'description'"

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})

	t.Run("Given no brand field, should return true and appropriate message", func(t *testing.T) {
		var req map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "A Red Jumper",
		}

		missing, message := MissingMandatoryClothingField(req)

		if !missing {
			t.Errorf("Expected missing to be true")
		}

		expectedMessage := "Invalid request body, missing 'brand'"

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})

	t.Run("Given no store field, should return true and appropriate message", func(t *testing.T) {
		var req map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "A Red Jumper",
			"brand":        "A&B",
		}

		missing, message := MissingMandatoryClothingField(req)

		if !missing {
			t.Errorf("Expected missing to be true")
		}

		expectedMessage := "Invalid request body, missing 'store'"

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})

	t.Run("Given no size field, should return true and appropriate message", func(t *testing.T) {
		var req map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "A Red Jumper",
			"brand":        "A&B",
			"store":        "My Real Store",
		}

		missing, message := MissingMandatoryClothingField(req)

		if !missing {
			t.Errorf("Expected missing to be true")
		}

		expectedMessage := "Invalid request body, missing 'size'"

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})

	t.Run("Given all required fields, should return false and empty message", func(t *testing.T) {
		var req map[string]any = map[string]any{
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "A Red Jumper",
			"brand":        "A&B",
			"store":        "My Real Store",
			"size":         "M",
		}

		missing, message := MissingMandatoryClothingField(req)

		if missing {
			t.Errorf("Expected missing to be false")
		}

		expectedMessage := ""

		if message != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message)
		}
	})
}

func TestCreateClothing(t *testing.T) {
	t.Run("Given request method is not allowed, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodTrace, "/clothes", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected a %d error, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unauthorised method %s.", http.MethodTrace)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, with no userId provided, should return an error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPost, "/clothes", http.NoBody)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected a %d error, got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expectedMessage := "UserID not provided"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expecteed %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, with no or empty body, should return an error", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", http.NoBody)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

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

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", strings.NewReader("data"))

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must be JSON"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has missing field, should return error", func(t *testing.T) {
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

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", bytes.NewReader(body))

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'pricePence'"

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

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", bytes.NewReader(body))

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

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

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", bytes.NewReader(body))

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

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

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", bytes.NewReader(body))

		repo := repository.NewInMemoryClothingRepository()
		repo.SetIssueOnSaveError(true)
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

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

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes", bytes.NewReader(body))

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.CreateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected %d got %d", http.StatusCreated, resp.StatusCode)
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

		if data["id"] == "" {
			t.Errorf("Expected id to be populated")
		}

		if data["clothingType"] != jsonMap["clothingType"] {
			t.Errorf("Expected returned clothingType %s, got %v", jsonMap["clothingType"], data["clothingType"])
		}

		if data["userId"] != "test-user-id" {
			t.Errorf("Expected returned userId %s, got %v", "test-user-id", data["userId"])
		}

	})
}

func TestGetClothing(t *testing.T) {

	t.Run("Given request method is not allowed, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodTrace, "/clothes", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.GetClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected a %d error, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unauthorised method %s.", http.MethodTrace)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given GET request, but no userID provided, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.GetClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected a %d error, got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expectedMessage := "UserID not provided"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given GET request, with issues with retrieval, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes", nil)

		repo := repository.NewInMemoryClothingRepository()
		repo.SetIssueOnGetAllError(true)
		apiHandler := &API{
			Repo: repo,
		}

		apiHandler.GetClothing(w, r)

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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes", nil)

		expectedItems := []domain.Clothing{
			{Id: "id-1", ClothingType: "Shirt", Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M"},
			{Id: "id-2", ClothingType: "Trousers", Description: "Black Jeans", Brand: "Y", Store: "Store", Price: 2500, Size: "L"},
		}
		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}

		for _, item := range expectedItems {
			repo.Save("test-user-id", item)
		}

		apiHandler.GetClothing(w, r)

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

func TestGetClothingById(t *testing.T) {

	t.Run("Given request method is not allowed, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodTrace, "/clothes/legit-id", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected a %d error, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unauthorised method %s.", http.MethodTrace)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given GET request, but no userId provided, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes/legit-id", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected a %d error, got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expectedMessage := "UserID not provided"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given missing id, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes/", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Missing 'id' parameter"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given empty or whitespace id, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes/", nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": ""})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Missing 'id' parameter"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given GET request, with issues with retrieval, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		repo := repository.NewInMemoryClothingRepository()
		repo.SetIssueOnGetByIdError(true)
		apiHandler := &API{
			Repo: repo,
		}
		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes/"+item.Id, nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": item.Id})

		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expeted %d got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		expected := fmt.Sprintf("Unable to get clothing for ID %s", item.Id)

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}

	})

	t.Run("Given GET request, with no issues with retrieval and item does not exists, should  return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes/legit-id", nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": "legit-id"})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}

		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expeted %d got %d", http.StatusNotFound, resp.StatusCode)
		}

		expected := "Clothing item not found for ID legit-id"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given GET request, with no issues with retrieval and item exists, should not return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt", Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/clothes/"+item.Id, nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": item.Id})

		apiHandler := &API{
			Repo: repo,
		}

		apiHandler.GetClothingById(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expeted %d got %d", http.StatusOK, resp.StatusCode)
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

		if data["id"] != item.Id {
			t.Errorf("Expected returned ID legit-id, got %v", data["id"])
		}
	})
}

func TestUpdateClothing(t *testing.T) {
	t.Run("Given request method is not allowed, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodTrace, "/clothes/test-id", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected a %d error, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unauthorised method %s.", http.MethodTrace)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method PUT, but no userId provided, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPut, "/clothes/test-id", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected a %d error, got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expectedMessage := "UserID not provided"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, with empty or whitespace id, should return an error", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/test-id", nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": ""})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Missing 'id' parameter"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, with valid id, but missing or empty body, should return an error", func(t *testing.T) {
		w := httptest.NewRecorder()

		id := "test-id"

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must not be empty or missing"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, where data isn't json, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		id := "test-id"

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")

		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, strings.NewReader("data"))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must be JSON"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, where data has missing field, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		id := "test-id"
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, missing 'pricePence'"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, where data has extra field, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		id := "test-id"
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, superfluous fields"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, where URL and body ID do not match, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		bodyId := "test-id"
		urlId := bodyId + uuid.New().String()
		var jsonMap map[string]any = map[string]any{
			"id":           bodyId,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+urlId, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": urlId})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Body has Id = %s, but id = %s, resulting is mismatch", bodyId, urlId)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, where clothes userId and userId do not match, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		bodyUserId := "test-user-id"
		suppliedUserId := bodyUserId + uuid.New().String()
		itemId := uuid.New().String()
		var jsonMap map[string]any = map[string]any{
			"id":           itemId,
			"pricePence":   2000,
			"clothingType": "Jumper",
			"description":  "Red loosefit jumper",
			"brand":        "A&B",
			"store":        "Totlly Real Store",
			"size":         "Medium",
			"userId":       bodyUserId,
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}
		ctx := context.WithValue(context.TODO(), UserIDContextKey, suppliedUserId)
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+itemId, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": itemId})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Body has UserId = %s, but UserId = %s, resulting is mismatch", bodyUserId, suppliedUserId)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, where data is structurally valid, but breaks validation, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		id := "test-id"
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, breaks validation rule:"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given PUT request, with valid data but issue on save, should have error", func(t *testing.T) {
		w := httptest.NewRecorder()

		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		id := item.Id
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})

		repo.SetIssueOnUpdateError(true)
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected %d got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		expectedMessage := "Error updating clothing item"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}

	})

	t.Run("Given PUT request, but ID does not exist, should have error", func(t *testing.T) {
		w := httptest.NewRecorder()

		id := "test-id"
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})
		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected %d got %d", http.StatusNotFound, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Clothing item not found for ID %s", id)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}

	})

	t.Run("Given PUT request, with valid data and no issue on save, should have success", func(t *testing.T) {
		w := httptest.NewRecorder()

		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		id := item.Id
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})
		expectedID := id

		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &responseBody)

		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v, %s", err, w.Body.String())
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

		if data["userId"] != "test-user-id" {
			t.Errorf("Expected returned userId %s, got %v", "test-user-id", data["userId"])
		}

	})

	t.Run("Given PATCH request, with valid data and no issue on save, should have success", func(t *testing.T) {
		w := httptest.NewRecorder()

		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		id := item.Id
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPatch, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})
		expectedID := id

		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &responseBody)

		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v, %s", err, w.Body.String())
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

		if data["userId"] != "test-user-id" {
			t.Errorf("Expected returned userId %s, got %v", "test-user-id", data["userId"])
		}

	})

	t.Run("Given POST request, with valid data and no issue on save, should have success", func(t *testing.T) {
		w := httptest.NewRecorder()

		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		id := item.Id
		var jsonMap map[string]any = map[string]any{
			"id":           id,
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
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/clothes/"+id, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": id})
		expectedID := id

		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.UpdateClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected %d got %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &responseBody)

		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v, %s", err, w.Body.String())
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

		if data["userId"] != "test-user-id" {
			t.Errorf("Expected returned userId %s, got %v", "test-user-id", data["userId"])
		}
	})
}

func TestDeleteClothing(t *testing.T) {
	t.Run("Given request method is not allowed, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodTrace, "/clothes/test-id", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected a %d error, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unauthorised method %s.", http.MethodTrace)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given DELETE request, but no userId provided, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodDelete, "/clothes/test-id", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected a %d error, got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expectedMessage := "UserID not provided"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given missing id, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/clothes/", nil)
		r.Header.Set("Content-Type", "application/json")

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Missing 'id' parameter"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given empty or whitespace id, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/clothes/", nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": ""})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}
		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected a %d error, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Missing 'id' parameter"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s, got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given DELETE request, with issues with delete, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/clothes/"+item.Id, nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": item.Id})

		repo.SetIssueOnDeleteError(true)
		apiHandler := &API{
			Repo: repo,
		}

		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expeted %d got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		expected := "Unable to delete clothing for ID " + item.Id

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}

	})

	t.Run("Given DELETE request, when id does not exist, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/clothes/legit-id", nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": "legit-id"})

		repo := repository.NewInMemoryClothingRepository()
		apiHandler := &API{
			Repo: repo,
		}

		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expeted %d got %d", http.StatusNotFound, resp.StatusCode)
		}

		expected := fmt.Sprintf("Clothing item not found for ID %s", "legit-id")

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}

	})

	t.Run("Given DELETE request, with no issues with delete, should not return error", func(t *testing.T) {
		w := httptest.NewRecorder()
		repo := repository.NewInMemoryClothingRepository()

		item := domain.Clothing{
			ClothingType: "Shirt " + uuid.New().String(), Description: "Blue Shirt", Brand: "X", Store: "Shop", Price: 1000, Size: "M",
		}

		item, err := repo.Save("test-user-id", item)

		if err != nil {
			t.Fatalf("Expected no err got %v", err)
		}

		ctx := context.WithValue(context.TODO(), UserIDContextKey, "test-user-id")
		r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/clothes/"+item.Id, nil)
		r.Header.Set("Content-Type", "application/json")
		r = mux.SetURLVars(r, map[string]string{"id": item.Id})

		apiHandler := &API{
			Repo: repo,
		}

		apiHandler.DeleteClothing(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expeted %d got %d", http.StatusOK, resp.StatusCode)
		}
	})
}
