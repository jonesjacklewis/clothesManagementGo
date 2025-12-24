package main

import (
	"clothes_management/internal/api"
	"clothes_management/internal/repository"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := 8080
	portStr := fmt.Sprintf(":%d", port)

	repo := repository.NewInMemoryClothingRepository()
	apiHandler := &api.API{
		Repo: repo,
	}

	http.HandleFunc("/clothes", apiHandler.Clothes)

	log.Fatal(http.ListenAndServe(portStr, nil))
}
