package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BookStore struct {
	db *sql.DB
}

func newBookStore(db *sql.DB) *BookStore {
	return &BookStore{db}
}

func (store *BookStore) GetBook(id uuid.UUID) (b Book, e error) {

	statement, err := store.db.Prepare(`
		select b.id, b.name, b.genre, b.created_at, b.publication_date, a.id, a.name, a.created_at
		from books b 
		inner join authors a on b.author_id=a.id where b.id=$1
	`)

	if err != nil {
		log.Println("BookStore.GetBook() - received error from db", err)
		return b, err
	}

	if scanErr := statement.QueryRow(id.String()).Scan(&b.Id, &b.Name, &b.Genre, &b.CreatedAt, &b.PublicationDate, &b.Author.Id, &b.Author.Name, &b.Author.CreatedAt); scanErr != nil {
		log.Println("BookStore.GetBook() - received error from db", scanErr)
		return b, scanErr
	}

	log.Println("BookStore.GetBook() - received from db", b)
	return b, nil
}

func (store *BookStore) GetBooks(m map[string]string) ([]Book, error) {
	query := "select id, name, genre, publication_date, created_at, author_id from books"
	params := make([]any, len(m))

	if len(m) != 0 {
		query += " where "
		var i = 1
		for k, v := range m {
			if k == "publication_date" {
				query += k + "=$" + fmt.Sprint(i) + " "
			} else {
				query += k + " like '%' || $" + fmt.Sprint(i) + " || '%' "
			}
			params[i-1] = v
			i++
		}
	}

	query = strings.TrimSpace(query)
	log.Println("BookStore.GetBooks() - executing query", query, params)

	statement, err := store.db.Prepare(query)

	if err != nil {
		log.Println("BookStore.GetBooks() - received error from db", err)
		return nil, err
	}

	var queryRows *sql.Rows
	var queryError error
	if len(params) == 0 {
		queryRows, queryError = statement.Query()
	} else {
		queryRows, queryError = statement.Query(params...)
	}

	if queryError != nil {
		log.Println("BookStore.GetBooks() - received error from db", queryError)
		return nil, queryError
	}

	var books []Book
	defer queryRows.Close()
	for queryRows.Next() {
		var book Book
		if scanErr := queryRows.Scan(&book.Id, &book.Name, &book.Genre, &book.PublicationDate, &book.CreatedAt, &book.Author.Id); scanErr != nil {
			log.Println("BookStore.GetBooks() - received error while scanning", err)
			return nil, err
		}

		books = append(books, book)
	}

	if err := queryRows.Err(); err != nil {
		log.Println("BookStore.GetBooks() - received error from db", err)
		return nil, err
	}

	return books, nil
}

func (store *BookStore) Remove(id uuid.UUID) error {

	statement, err := store.db.Prepare(`delete from books where id=$1`)

	if err != nil {
		log.Println("BookStore.Remove() received error from db", err)
		return err
	}

	if _, execErr := statement.Exec(id); execErr != nil {
		log.Println("BookStore.Remove() received error from db", execErr)
		return execErr
	}

	return nil
}

func (store *BookStore) CreateBook(b Book) (savedBook Book, err error) {

	statement, err := store.db.Prepare(`
		with new_book as (	
			insert into books(name, genre, publication_date, created_at, author_id)
				select $1, $2, $3, $4, authors.id from authors where authors.id=$5 
				returning *
		)
		select new_book.*, authors.name, authors.created_at from new_book inner join authors on new_book.author_id = authors.id
	`)

	if err != nil {
		log.Println("BookStore.CreateBook() received error from db", err)
		return b, err
	}

	row := statement.QueryRow(&b.Name, &b.Genre, &b.PublicationDate, time.Now().UTC(), &b.Author.Id)

	scanError := row.Scan(&savedBook.Id, &savedBook.Name, &savedBook.Genre, &savedBook.PublicationDate,
		&savedBook.CreatedAt, &savedBook.Author.Id, &savedBook.Author.Name, &savedBook.Author.CreatedAt)

	if scanError != nil {
		log.Println("BookStore.CreateBook() received error from db", scanError)
		return b, err
	}

	savedBook.Author.Id = b.Author.Id

	return savedBook, nil
}

func (store *BookStore) UpdateBook(b Book) (updatedBook Book, err error) {

	statement, err := store.db.Prepare(`
		with updated_book as (
			update books set name=$1, genre=$2, publication_date=$3, author_id=$4 returning *
		)
		select updated_book.*, authors.name, authors.created_at from updated_book
		inner join authors on updated_book.author_id = authors.id
	`)

	if err != nil {
		log.Println("BookStore.UpdateBook() received error from db", err)
		return b, err
	}

	row := statement.QueryRow(&b.Name, &b.Genre, &b.PublicationDate, &b.Author.Id)

	scanError := row.Scan(
		&updatedBook.Id, &updatedBook.Name, &updatedBook.Genre,
		&updatedBook.PublicationDate, &updatedBook.CreatedAt,
		&updatedBook.Author.Id, &updatedBook.Author.Name, &updatedBook.Author.CreatedAt)

	if scanError != nil {
		log.Println("BookStore.UpdateBook() received error from db", scanError)
		return b, err
	}

	return updatedBook, nil
}
