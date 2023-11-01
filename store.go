package main

import (
	"github.com/google/uuid"
)

type BookStore interface {
	GetBook(id uuid.UUID) (b Book, e error)
	GetBooks() (b []Book, e error)
	Remove(id uuid.UUID) error
	CreateBook(b Book) error
	UpdateBook(b Book) error
}
