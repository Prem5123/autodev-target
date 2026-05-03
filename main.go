package main

import (
	"log"
	"net/http"

	"github.com/Prem5123/autodev-target/api"
	"github.com/Prem5123/autodev-target/auth"
	"github.com/Prem5123/autodev-target/notes"
)

func main() {
	// Initialize storage
	storage, err := notes.NewFileStorage("notes.json")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Initialize note service
	noteService := notes.NewService(storage)

	// Initialize auth service
	authService := auth.NewAuthService()

	// Set up routes
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/api/auth/login", authService.LoginHandler)
	mux.HandleFunc("/api/auth/logout", authService.LogoutHandler)

	// Protected API routes
	apiHandler := api.NewHandler(noteService, authService)
	apiHandler.RegisterRoutes(mux)

	// Serve static files from current directory
	mux.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}