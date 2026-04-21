package main

import "time"

type Request struct {
	Body string `json:"body"`
}

type ResponseValid struct {
	Valid bool `json:"valid"`
}

type ResponseClean struct {
	CleanedBody string `json:"cleaned_body"`
}

type ResponseChirp struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

type ResponseUser struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type CreateUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
