package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dvjn/sorcerer/internal/auth"
	"github.com/dvjn/sorcerer/internal/config"
	"github.com/dvjn/sorcerer/internal/storage"
	"github.com/dvjn/sorcerer/internal/web/controller"
	"github.com/dvjn/sorcerer/internal/web/router"
)

func main() {
	config := config.LoadConfig()

	storage, err := storage.New(&config.Storage)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}

	auth, err := auth.New(config)
	if err != nil {
		fmt.Printf("Failed to initialize auth: %v\n", err)
		os.Exit(1)
	}

	controller := controller.New(storage)

	router := router.New(auth, controller)

	fmt.Printf("Starting server on port %d\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
}
