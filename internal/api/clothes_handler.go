package api

import (
	"bytes"
	"clothes_management/internal/domain"
	"clothes_management/internal/repository"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type API struct {
	Repo repository.ClothingRepository
}

func (a *API) clothes_post(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil || r.Body == http.NoBody {
		http.Error(w, "Request body must not be empty or missing", http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req map[string]any

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Request body must be JSON", http.StatusBadRequest)
		return
	}

	_, exists := req["pricePence"]

	if !exists {
		http.Error(w, "Invalid request body, missing 'pricePence'", http.StatusBadRequest)
		return
	}

	_, exists = req["clothingType"]

	if !exists {
		http.Error(w, "Invalid request body, missing 'clothingType'", http.StatusBadRequest)
		return
	}

	_, exists = req["description"]

	if !exists {
		http.Error(w, "Invalid request body, missing 'description'", http.StatusBadRequest)
		return
	}

	_, exists = req["brand"]

	if !exists {
		http.Error(w, "Invalid request body, missing 'brand'", http.StatusBadRequest)
		return
	}

	_, exists = req["store"]

	if !exists {
		http.Error(w, "Invalid request body, missing 'store'", http.StatusBadRequest)
		return
	}

	_, exists = req["size"]

	if !exists {
		http.Error(w, "Invalid request body, missing 'size'", http.StatusBadRequest)
		return
	}

	var clothing domain.Clothing

	dec := json.NewDecoder(bytes.NewReader(bodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&clothing); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body, superfluous fields %s", err.Error()), http.StatusBadRequest)
		return
	}

	err = clothing.Validate()

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body, breaks validation rule: %s", err.Error()), http.StatusBadRequest)
		return
	}

	clothing, err = a.Repo.Save(clothing)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error saving clothing item", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"success": true, "data": clothing}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(resp)

}

func (a *API) clothes_get(w http.ResponseWriter) {
	clothingItems, err := a.Repo.GetAll()

	if err != nil {
		log.Print(err)
		http.Error(w, "Error getting clothing items", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"success": true, "data": clothingItems}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(resp)
}

func (a *API) Clothes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unauthorised method %s. The supported methods are POST or GET.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodPost {
		a.clothes_post(w, r)
		return
	} else {
		a.clothes_get(w)
		return
	}

}
