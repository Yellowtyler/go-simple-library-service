package main

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	Id        uuid.UUID `json:id`
	Name      string    `json:Name`
	CreatedAt time.Time `json:createdAt`
	Genre     string    `json:genre`
	Author    Author    `json:author`
}

type Author struct {
	Id        uuid.UUID `json:id`
	Name      string    `json:Name`
	CreatedAt time.Time `json:createdAt`
}
