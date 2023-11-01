package main

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
)

type BookStore struct {
	db *sql.DB
}

func newBookStore(db *sql.DB) *BookStore {
	return &BookStore{db}
}

func (store *BookStore) GetBook(id uuid.UUID) (b Book, e error) {

	row := store.db.QueryRow(`select b.id, b.name, b.genre, b.created_at, a.id, a.name, a.created_at 
		from books b 
		inner join authors a on b.author_id=a.id where b.id=$1`, id.String())

	var book Book
	var author Author
	if scanErr := row.Scan(&book.Id, &book.Name, &book.Genre, &book.CreatedAt, &author.Id, &author.Name, &author.CreatedAt); scanErr != nil {
		log.Println(scanErr)
		return book, scanErr
	}

	book.Author = author

	log.Println("GetBook() - received from db", book)
	return book, nil
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
