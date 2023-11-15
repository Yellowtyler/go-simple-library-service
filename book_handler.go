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
	s *BookStore
}

func NewBookHandler(db *sql.DB) *BookHandler {
	store := NewBookStore(db)
	return &BookHandler{store}
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
	var token = r.Header.Get("Authorization")

	if token == "" {
		return
	}

	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("BookHandler.getBook() - processing request", r.URL.Path)

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("BookHandler.getBook() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	var book Book
	if book, err = BookHandler.s.GetBook(id); err != nil {
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

	if !ValidParams("book", queryMap) {
		log.Println("BookHandler.getBooks() - received invalid params!", queryMap)

		HandleError(500, "Internal Server Error", w)
		return
	}

	var books []Book
	var err error
	if books, err = BookHandler.s.GetBooks(queryMap); err != nil {
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

	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Println("BookHandler.createBook() - received decode error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.createBook() - received req", book)

	var savedBook Book
	var err error
	if savedBook, err = BookHandler.s.CreateBook(book); err != nil {
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
	log.Println("BookHandler.updateBook() - received req", r.Body)

	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Println("BookHandler.updateBook() - received decode error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	var updatedBook Book
	var err error
	if updatedBook, err = BookHandler.s.UpdateBook(book); err != nil {
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

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("deleteBook() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	if err = BookHandler.s.Remove(id); err != nil {
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
