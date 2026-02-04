package main

import (
	"log"
	"net/http"
	"os"

	"practice/internal/httpapi"
	"practice/internal/shortener"
)

func main() {
	baseURL := os.Getenv("BASE_URL")

	store := shortener.NewStore()
	service := shortener.NewService(store)
	handler := httpapi.NewHandler(service, baseURL)

	mux := http.NewServeMux()
	handler.Register(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
