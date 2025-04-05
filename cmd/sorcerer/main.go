package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dvjn/sorcerer/pkg/api"
	"github.com/dvjn/sorcerer/pkg/config"
	"github.com/dvjn/sorcerer/pkg/storage"
)

func main() {
	config := config.LoadConfig()

	storage, err := storage.NewStorage(config.StoragePath)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}

	handlers := api.NewHandlers(storage)

	router := api.SetupRouter(handlers)

	fmt.Printf("Starting server on port %s\n", config.Port)
	http.ListenAndServe(":"+config.Port, router)
}
