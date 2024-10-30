package domain

import (
	"time"
)

type CreateAccountRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type Account struct {
	ID        int       `json:"id"`
	UserName  string    `json:"userName"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewAccount(userName string, password string) *Account {
	return &Account{
		UserName:  userName,
		Password:  password,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
