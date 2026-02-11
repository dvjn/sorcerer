package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var jwksFile string

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	keyDir := os.Getenv("KEY_DIR")
	if keyDir == "" {
		keyDir = "test-keys"
	}

	jwksFile = filepath.Join(keyDir, "jwks.json")

	// Verify JWKS file exists
	if _, err := os.Stat(jwksFile); os.IsNotExist(err) {
		log.Fatalf("JWKS file not found: %s (run 'go run ./cmd/gen-test-jwt' first)", jwksFile)
	}

	// Setup HTTP handlers
	http.HandleFunc("/.well-known/jwks.json", jwksHandler)
	http.HandleFunc("/jwks.json", jwksHandler)
	http.HandleFunc("/health", healthHandler)

	addr := ":" + port
	log.Printf("JWKS server listening on %s", addr)
	log.Printf("Serving JWKS from: %s", jwksFile)
	log.Printf("Endpoints:")
	log.Printf("  http://localhost%s/.well-known/jwks.json", addr)
	log.Printf("  http://localhost%s/jwks.json", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func jwksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=300")

	data, err := os.ReadFile(jwksFile)
	if err != nil {
		http.Error(w, `{"error":"Failed to read JWKS"}`, http.StatusInternalServerError)
		return
	}

	// Validate it's valid JSON
	if !json.Valid(data) {
		http.Error(w, `{"error":"Invalid JWKS format"}`, http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
