package domain

import (
	"time"
)

type CreateRatingRequest struct {
	Rating int    `json:"rating"`
	Review string `json:"review"`
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
