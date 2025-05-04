package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dvjn/sorcerer/internal/auth"
	"github.com/dvjn/sorcerer/internal/config"
	"github.com/dvjn/sorcerer/internal/router"
	"github.com/dvjn/sorcerer/internal/service"
	"github.com/dvjn/sorcerer/internal/storage"
)

func main() {
	config := config.LoadConfig()

	storage, err := storage.NewStorage(config.StoragePath)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}

	auth := auth.NewAuth(config)

	service := service.NewService(storage)

	router := router.SetupRouter(service, auth)

	fmt.Printf("Starting server on port %d\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
}
