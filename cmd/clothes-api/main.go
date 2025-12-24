package main

import (
	"clothes_management/internal/api"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := 8080
	portStr := fmt.Sprintf(":%d", port)

	http.HandleFunc("/clothes", api.Clothes)

	log.Fatal(http.ListenAndServe(portStr, nil))
}
