package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello Go Gym Management!")

	store, err := NewPostgreSQLStore()

	if err != nil {
		log.Fatal("Failed to create DB store")
	}

	fmt.Printf("%v\n", store)

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize DB store %s", err.Error())
	}

	server := NewAPIServer(":8000", store)
	server.Run()
}
