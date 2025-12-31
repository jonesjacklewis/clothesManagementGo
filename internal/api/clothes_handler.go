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
	"slices"
	"strings"

	"github.com/gorilla/mux"
)

type API struct {
	Repo repository.ClothingRepository
}

func (a *API) CreateClothing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unauthorised method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

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

	missingField, message := MissingMandatoryClothingField(req)

	if missingField {
		http.Error(w, message, http.StatusBadRequest)
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

func (a *API) GetClothing(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Unauthorised method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

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

func (a *API) GetClothingById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Unauthorised method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, exists := vars["id"]

	if !exists {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	id = strings.TrimSpace(id)

	if len(id) == 0 {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	item, err := a.Repo.GetById(id)

	if err != nil {
		log.Print(err)
		http.Error(w, fmt.Sprintf("Unable to get clothing for ID %s", id), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"success": true, "data": item}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(resp)
}

func (a *API) UpdateClothing(w http.ResponseWriter, r *http.Request) {
	allowedMethods := []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
	}

	if !slices.Contains(allowedMethods, r.Method) {
		http.Error(w, fmt.Sprintf("Unauthorised method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, exists := vars["id"]

	if !exists {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	id = strings.TrimSpace(id)

	if len(id) == 0 {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

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

	missingField, message := MissingMandatoryClothingField(req)

	if missingField {
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	var clothing domain.Clothing

	dec := json.NewDecoder(bytes.NewReader(bodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&clothing); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body, superfluous fields %s", err.Error()), http.StatusBadRequest)
		return
	}

	if clothing.Id != "" && clothing.Id != id {
		http.Error(w, fmt.Sprintf("Body has Id = %s, but id = %s, resulting is mismatch", clothing.Id, id), http.StatusBadRequest)
		return
	}

	if clothing.Id == "" {
		clothing.Id = id
	}

	err = clothing.Validate()

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body, breaks validation rule: %s", err.Error()), http.StatusBadRequest)
		return
	}

	clothing, err = a.Repo.Update(clothing)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error updating clothing item", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"success": true, "data": clothing}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(resp)

}

func (a *API) DeleteClothing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, fmt.Sprintf("Unauthorised method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, exists := vars["id"]

	if !exists {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	id = strings.TrimSpace(id)

	if len(id) == 0 {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	err := a.Repo.Delete(id)

	if err != nil {
		log.Print(err)
		http.Error(w, fmt.Sprintf("Unable to delete clothing for ID %s", id), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"success": true, "id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	json.NewEncoder(w).Encode(resp)
}

func MissingMandatoryClothingField(req map[string]any) (bool, string) {

	var requiredFields []string = []string{
		"pricePence",
		"clothingType",
		"description",
		"brand",
		"store",
		"size",
	}

	for _, requiredField := range requiredFields {
		_, exists := req[requiredField]
		if !exists {
			return true, fmt.Sprintf("Invalid request body, missing '%s'", requiredField)
		}

	}

	return false, ""
}
