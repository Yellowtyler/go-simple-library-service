package entity

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	Id              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	PublicationDate string    `json:"publicationDate"`
	CreatedAt       time.Time `json:"createdAt"`
	Genre           string    `json:"genre"`
	Author          Author    `json:"author"`
}
