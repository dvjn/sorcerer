package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dvjn/sorcerer/pkg/auth"
	"github.com/dvjn/sorcerer/pkg/config"
	"github.com/dvjn/sorcerer/pkg/router"
	"github.com/dvjn/sorcerer/pkg/service"
	"github.com/dvjn/sorcerer/pkg/storage"
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
