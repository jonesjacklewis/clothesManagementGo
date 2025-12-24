package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClothes(t *testing.T) {
	t.Run("Given request doesn't use POST or GET, should return error", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodDelete, "/clothes", nil)
		r.Header.Set("Content-Type", "application/json")

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

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

		Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Invalid request body, breaks validation rule:"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given POST request, with valid data, should return StatusCreated", func(t *testing.T) {
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

		Clothes(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected %d got %d", http.StatusCreated, resp.StatusCode)
		}

	})

}
