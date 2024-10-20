package main

import "fmt"

func main() {
	fmt.Println("Hello Go Gym Management!")

	server := NewAPIServer(":8000")
	server.Run()
}
