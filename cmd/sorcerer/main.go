package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dvjn/sorcerer/internal/auth"
	"github.com/dvjn/sorcerer/internal/config"
	"github.com/dvjn/sorcerer/internal/store"
	"github.com/dvjn/sorcerer/internal/web/controller"
	"github.com/dvjn/sorcerer/internal/web/router"
)

func main() {
	config, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	errors := config.Validate()
	if len(errors) > 0 {
		fmt.Println("Config validations failed:")
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		os.Exit(1)
	}

	store, err := store.New(&config.Store)
	if err != nil {
		fmt.Printf("Failed to initialize store: %v\n", err)
		os.Exit(1)
	}

	auth, err := auth.New(config)
	if err != nil {
		fmt.Printf("Failed to initialize auth: %v\n", err)
		os.Exit(1)
	}

	controller := controller.New(store)

	router := router.New(auth, controller)

	fmt.Printf("Starting server on port %d\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
}
