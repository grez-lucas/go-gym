package domain

import (
	"time"
)

type CreateGymRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Gym struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Rating      float32   `json:"rating"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewGym(name string, description string) *Gym {
	return &Gym{
		Name:        name,
		Description: description,
		Rating:      0,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}
