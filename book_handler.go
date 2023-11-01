package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type BookHandler struct {
	s *BookStore
}

func newBookHandler(db *sql.DB) *BookHandler {
	store := newBookStore(db)
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
	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("getBook() - processing request", r.URL.Path)

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("getBook() - received error", err)
		InternalServerErrorHandler(w, r)
		return
	}

	var book Book
	if book, err = BookHandler.s.GetBook(id); err != nil {
		log.Println("getBook() - received error from db", err)
		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(book)
	if err != nil {
		log.Println("getBook() - received error while marshaling", err)
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (BookHandler *BookHandler) getBooks(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get books"))

}

func (BookHandler *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		InternalServerErrorHandler(w, r)
	}

}

func (BookHandler *BookHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("update book"))

}

func (BookHandler *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("delete book"))
}
