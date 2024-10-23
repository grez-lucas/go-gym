package main

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

type CreateRatingRequest struct {
	Rating   int    `json:"rating"`
	UserName string `json:"userName"`
	Review   string `json:"review"`
}

type Rating struct {
	ID        int       `json:"id"`
	GymID     int       `json:"gymId"`
	Rating    int       `json:"rating"`
	UserName  string    `json:"userName"`
	Review    string    `json:"review"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewRating(gymID int, rating int, userName string, review string) *Rating {
	return &Rating{
		GymID:     gymID,
		Rating:    rating,
		UserName:  userName,
		Review:    review,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
