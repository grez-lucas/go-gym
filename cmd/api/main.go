package main

import (
	"fmt"
	"log"

	"github.com/grez-lucas/go-gym/pkg/http"
	"github.com/grez-lucas/go-gym/pkg/storage"
)

func main() {
	fmt.Println("Hello Go Gym Management!")

	store, err := storage.NewPostgreSQLStore()

	if err != nil {
		log.Fatal("Failed to create DB store ", err.Error())
	}

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize DB store %s", err.Error())
	}

	server := http.NewAPIServer(":8000", store)
	server.Run()
}

// TODO: Refactor app structure
// TODO: Add go commands to purge DB (dropping tables)
