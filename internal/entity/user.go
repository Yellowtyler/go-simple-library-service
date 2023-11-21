package entity

import "github.com/google/uuid"

const (
	USER      = iota
	MODERATOR = iota
	ADMIN     = iota
)

type User struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Mail      string    `json:"mail"`
	Role      int       `json:"role"`
	Password  string    `json:"-"`
	CreatedAt string    `json:"createdAt"`
	Token     string    `json:"-"`
}
