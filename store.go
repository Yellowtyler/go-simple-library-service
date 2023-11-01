package main

import (
	"database/sql"

	"github.com/google/uuid"
)

type BookStore struct {
	db *sql.DB
}

func newBookStore(db *sql.DB) *BookStore {
	return &BookStore{db}
}

func (store *BookStore) GetBook(id uuid.UUID) (b Book, e error) {
	return
}
func (store *BookStore) GetBooks() (b []Book, e error) {
	return
}

func (store *BookStore) Remove(id uuid.UUID) error {
	return nil
}

func (store *BookStore) CreateBook(b Book) error {
	return nil
}

func (store *BookStore) UpdateBook(b Book) error {
	return nil
}
