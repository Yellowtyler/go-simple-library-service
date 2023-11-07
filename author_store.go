package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthorStore struct {
	db *sql.DB
}

func NewAuthorStore(db *sql.DB) *AuthorStore {
	return &AuthorStore{db}
}

func (store *AuthorStore) GetAuthor(id uuid.UUID) (a Author, e error) {
	statement, err := store.db.Prepare(`
		select a.*, b.id, b.name, b.genre, b.publication_date, b.created_at
		from authors a 
		join books b on a.id=b.author_id where a.id=$1
	`)

	if err != nil {
		log.Println("AuthorStore.GetAuthor() - received error from db", err)
		return a, err
	}

	rows, rowErr := statement.Query(id.String())

	if rowErr != nil {
		log.Println("AuthorStore.GetAuthor() - received error from db", rowErr)
		return a, rowErr
	}

	var books []AuthorBook
	for rows.Next() {
		var book AuthorBook
		if scanErr := rows.Scan(&a.Id, &a.Name, &a.CreatedAt, &book.Id, &book.Name, &book.Genre, &book.PublicationDate, &book.CreatedAt); scanErr != nil {
			log.Println("AuthorStore.GetAuthor() - received error from db", scanErr)
			return a, scanErr
		}

		books = append(books, book)
	}

	a.Books = books
	log.Println("AuthorStore.GetAuthor() - received from db", a)
	return a, nil
}

func (store *AuthorStore) GetAuthors(m map[string]string) ([]Author, error) {
	query := `select a.*, b.id, b.name, b.genre, b.publication_date, b.created_at from authors a 
		left join books b on a.id = b.author_id`
	params := make([]any, len(m))

	if len(m) != 0 {
		query += " where "
		var i = 1
		for k, v := range m {
			var column string
			switch {
			case k == "author_name":
				column = "a.name"
			case k == "book_name":
				column = "b.name"
			default:
				column = "a." + k
			}

			query += column + " like '%' || $" + fmt.Sprint(i) + " || '%' and "
			params[i-1] = v
			i++
		}
		query = strings.TrimSpace(query)
		queryArr := strings.Split(query, " ")
		query = strings.Join(queryArr[:len(queryArr)-1], " ")
	}

	log.Println("AuthorStore.GetAuthors() - executing query", query, params)

	statement, err := store.db.Prepare(query)

	if err != nil {
		log.Println("AuthorStore.GetAuthors() - received error from db", err)
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
		log.Println("AuthorStore.GetAuthors() - received error from db", queryError)
		return nil, queryError
	}

	authorsMap := make(map[uuid.UUID]Author)
	booksMap := make(map[uuid.UUID][]AuthorBook)
	defer queryRows.Close()
	for queryRows.Next() {
		var author Author
		var book AuthorBook
		if scanErr := queryRows.Scan(&author.Id, &author.Name, &author.CreatedAt, &book.Id, &book.Name, &book.Genre, &book.PublicationDate, &book.CreatedAt); scanErr != nil {
			log.Println("AuthorStore.GetAuthors() - received error while scanning", err)
		}

		books, ok := booksMap[author.Id]
		if !ok {
			books = make([]AuthorBook, 0)
		}

		if book.Id != uuid.Nil {
			books = append(books, book)
		}

		booksMap[author.Id] = books
		authorsMap[author.Id] = author
	}

	if err := queryRows.Err(); err != nil {
		log.Println("AuthorStore.GetAuthors() - received error from db", err)
		return nil, err
	}

	authors := make([]Author, 0)
	for _, v := range authorsMap {
		authors = append(authors, v)
	}

	for i := range authors {
		books := booksMap[authors[i].Id]
		authors[i].Books = books
	}

	return authors, nil
}

func (store *AuthorStore) CreateAuthor(author Author) (savedAuthor Author, err error) {
	statement, err := store.db.Prepare(`
		insert into authors(name, created_at)
			values($1, $2) 
			returning *
	`)

	if err != nil {
		log.Println("AuthorStore.CreateAuthor() - received error from db", err)
		return savedAuthor, err
	}

	row := statement.QueryRow(&author.Name, time.Now().UTC())

	scanError := row.Scan(&savedAuthor.Id, &savedAuthor.Name, &savedAuthor.CreatedAt)

	if scanError != nil {
		log.Println("AuthorStore.CreateAuthor() - received error from db", scanError)
		return savedAuthor, scanError
	}

	return savedAuthor, nil
}

func (store *AuthorStore) UpdateAuthor(author Author) (updatedAuthor Author, err error) {
	statement, err := store.db.Prepare(`
		update authors set name=$1 where id=$2
		returning *
	`)

	if err != nil {
		log.Println("AuthorStore.UpdateAuthor() - received error from db", err)
		return updatedAuthor, err
	}

	row := statement.QueryRow(&author.Name, &author.Id)

	if scanError := row.Scan(&updatedAuthor.Id, &updatedAuthor.Name, &updatedAuthor.CreatedAt); scanError != nil {
		log.Println("AuthorStore.UpdateAuthor() - received error from db", scanError)
		return updatedAuthor, scanError
	}

	return updatedAuthor, nil
}

func (store *AuthorStore) DeleteAuthor(id uuid.UUID) error {

	updateBooksStatement, updateErr := store.db.Prepare("update books set author_id=null where author_id=$1")

	if updateErr != nil {
		log.Println("AuthorStore.DeleteAuthor() - received error from db", updateErr)
		return updateErr
	}

	deleteStatement, deleteErr := store.db.Prepare(`delete from authors where id=$1`)

	if deleteErr != nil {
		log.Println("AuthorStore.DeleteAuthor() - received error from db", deleteErr)
		return deleteErr
	}

	if _, execErr := updateBooksStatement.Exec(id); execErr != nil {
		log.Println("AuthorStore.DeleteAuthor() - received error from db", execErr)
		return execErr
	}

	if _, execErr := deleteStatement.Exec(id); execErr != nil {
		log.Println("AuthorStore.DeleteAuthor() - received error from db", execErr)
		return execErr
	}

	return nil
}
