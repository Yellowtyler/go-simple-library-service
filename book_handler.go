package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type BookHandler struct {
	S         *BookStore
	UserStore *UserStore
}

func NewBookHandler(db *sql.DB, userStore *UserStore) *BookHandler {
	store := NewBookStore(db)
	return &BookHandler{store, userStore}
}

func (bookHandler *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && BookRe.Match([]byte(r.URL.Path)):
		bookHandler.getBooks(w, r)
		return
	case r.Method == http.MethodGet && BookReWithID.Match([]byte(r.URL.Path)):
		bookHandler.getBook(w, r)
		return
	case r.Method == http.MethodPost && BookRe.Match([]byte(r.URL.Path)):
		bookHandler.createBook(w, r)
		return
	case r.Method == http.MethodPut && BookRe.Match([]byte(r.URL.Path)):
		bookHandler.updateBook(w, r)
		return
	case r.Method == http.MethodDelete && BookReWithID.Match([]byte(r.URL.Path)):
		bookHandler.deleteBook(w, r)
		return
	default:
		return
	}
}

func (BookHandler *BookHandler) getBook(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("BookHandler.getBook() - processing request", r.URL.Path)

	if err = ValidateToken(r.Header.Get("Authorization"), BookHandler.UserStore); err != nil {
		log.Println("BookHandler.getBook() - invalid token", err)
		HandleError(401, err.Error(), w)
		return
	}

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("BookHandler.getBook() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	var book Book
	if book, err = BookHandler.S.GetBook(id); err != nil {
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("book with id %v wasn't found", id), w)
			return
		}
		log.Println("BookHandler.getBook() - received error from db", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(book)
	if err != nil {
		log.Println("BookHandler.getBook() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.getBook() - successfully finished req", book)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (BookHandler *BookHandler) getBooks(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	queryMap := ToMap(values)
	log.Println("BookHandler.getBooks() - received req", queryMap)
	var err error

	if err = ValidateToken(r.Header.Get("Authorization"), BookHandler.UserStore); err != nil {
		log.Println("BookHandler.getBooks() - invalid token", err)
		HandleError(401, err.Error(), w)
		return
	}

	if !ValidParams("book", queryMap) {
		log.Println("BookHandler.getBooks() - received invalid params!", queryMap)

		HandleError(400, "Invalid request params", w)
		return
	}

	var books []Book
	if books, err = BookHandler.S.GetBooks(queryMap); err != nil {
		log.Println("BookHandler.getBooks() - received error from db", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(books)
	if err != nil {
		log.Println("BookHandler.getBooks() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.getBooks() - successfully finished req", books)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (BookHandler *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var invoker User
	var err error
	if invoker, err = ValidateTokenAndGetUser(r.Header.Get("Authorization"), BookHandler.UserStore); err != nil {
		log.Println("BookHandler.createBook() - invalid token", err)
		HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		HandleError(403, "403 Forbidden", w)
		return
	}

	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Println("BookHandler.createBook() - received decode error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.createBook() - received req", book)

	var savedBook Book
	if savedBook, err = BookHandler.S.CreateBook(book); err != nil {
		log.Println("BookHandler.createBook() - received error from db", err)
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("author with id %v wasn't found", book.Author.Id), w)
			return
		}
		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(savedBook)
	if err != nil {
		log.Println("BookHandler.createBook() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.createBook() - successfully finished req", savedBook)
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (BookHandler *BookHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	var invoker User
	var err error
	if invoker, err = ValidateTokenAndGetUser(r.Header.Get("Authorization"), BookHandler.UserStore); err != nil {
		log.Println("BookHandler.updateBook() - invalid token", err)
		HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		HandleError(403, "403 Forbidden", w)
		return
	}

	log.Println("BookHandler.updateBook() - received req", r.Body)

	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Println("BookHandler.updateBook() - received decode error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	var updatedBook Book
	if updatedBook, err = BookHandler.S.UpdateBook(book); err != nil {
		log.Println("BookHandler.updateBook() - received error from db", err)
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("book with id %v wasn't found", book.Id), w)
			return
		}
		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(updatedBook)
	if err != nil {
		log.Println("BookHandler.updateBook() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.updateBook() - successfully finished req", updatedBook)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (BookHandler *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("deleteBook() - processing request", r.URL.Path)

	var invoker User
	if invoker, err = ValidateTokenAndGetUser(r.Header.Get("Authorization"), BookHandler.UserStore); err != nil {
		log.Println("BookHandler.deleteBook() - invalid token", err)
		HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		HandleError(403, "403 Forbidden", w)
		return
	}
	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("deleteBook() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	if err = BookHandler.S.Remove(id); err != nil {
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("book with id %v wasn't found", id), w)
			return
		}

		log.Println("deleteBook() - received error from db", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.deleteBook() - successfully finished req", id)

	w.WriteHeader(http.StatusNoContent)
}
