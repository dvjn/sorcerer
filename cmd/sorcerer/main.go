package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dvjn/sorcerer/pkg/api"
	"github.com/dvjn/sorcerer/pkg/storage/filesystem"
)

func main() {
	storage, err := filesystem.NewFileSystemStorage("data")
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}

	handlers := api.NewHandlers(storage)

	router := api.SetupRouter(handlers)

	fmt.Println("Starting server on port 3000")
	http.ListenAndServe(":3000", router)
}
