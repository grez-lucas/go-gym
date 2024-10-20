package main

import "math/rand"

type Gym struct {
	ID          int
	Name        string
	Description string
	Rating      int
}

func NewGym(name string, description string) *Gym {
	return &Gym{
		ID:          rand.Intn(10000),
		Name:        name,
		Description: description,
		Rating:      rand.Intn(5),
	}
}
