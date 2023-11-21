package book

import (
	"database/sql"
	"encoding/json"
	"example/library-service/internal/auth"
	"example/library-service/internal/entity"
	"example/library-service/internal/errors"
	"example/library-service/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type BookHandler struct {
	bookStore *BookStore
	authStore *auth.AuthStore
}

func NewBookHandler(db *sql.DB, authStore *auth.AuthStore) *BookHandler {
	store := NewBookStore(db)
	return &BookHandler{store, authStore}
}

func (bookHandler *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && utils.BookRe.Match([]byte(r.URL.Path)):
		bookHandler.getBooks(w, r)
		return
	case r.Method == http.MethodGet && utils.BookReWithID.Match([]byte(r.URL.Path)):
		bookHandler.getBook(w, r)
		return
	case r.Method == http.MethodPost && utils.BookRe.Match([]byte(r.URL.Path)):
		bookHandler.createBook(w, r)
		return
	case r.Method == http.MethodPut && utils.BookRe.Match([]byte(r.URL.Path)):
		bookHandler.updateBook(w, r)
		return
	case r.Method == http.MethodDelete && utils.BookReWithID.Match([]byte(r.URL.Path)):
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

	if err = auth.ValidateToken(r.Header.Get("Authorization"), BookHandler.authStore); err != nil {
		log.Println("BookHandler.getBook() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("BookHandler.getBook() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	var book entity.Book
	if book, err = BookHandler.bookStore.GetBook(id); err != nil {
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("book with id %v wasn't found", id), w)
			return
		}
		log.Println("BookHandler.getBook() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(book)
	if err != nil {
		log.Println("BookHandler.getBook() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.getBook() - successfully finished req", book)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (BookHandler *BookHandler) getBooks(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	queryMap := utils.ToMap(values)
	log.Println("BookHandler.getBooks() - received req", queryMap)
	var err error

	if err = auth.ValidateToken(r.Header.Get("Authorization"), BookHandler.authStore); err != nil {
		log.Println("BookHandler.getBooks() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if !utils.ValidParams("book", queryMap) {
		log.Println("BookHandler.getBooks() - received invalid params!", queryMap)

		errors.HandleError(400, "Invalid request params", w)
		return
	}

	var books []entity.Book
	if books, err = BookHandler.bookStore.GetBooks(queryMap); err != nil {
		log.Println("BookHandler.getBooks() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(books)
	if err != nil {
		log.Println("BookHandler.getBooks() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.getBooks() - successfully finished req", books)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (BookHandler *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var invoker entity.User
	var err error
	if invoker, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), BookHandler.authStore); err != nil {
		log.Println("BookHandler.createBook() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != entity.MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	var book entity.Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Println("BookHandler.createBook() - received decode error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.createBook() - received req", book)

	var savedBook entity.Book
	if savedBook, err = BookHandler.bookStore.CreateBook(book); err != nil {
		log.Println("BookHandler.createBook() - received error from db", err)
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("author with id %v wasn't found", book.Author.Id), w)
			return
		}
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(savedBook)
	if err != nil {
		log.Println("BookHandler.createBook() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.createBook() - successfully finished req", savedBook)
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (BookHandler *BookHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	var invoker entity.User
	var err error
	if invoker, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), BookHandler.authStore); err != nil {
		log.Println("BookHandler.updateBook() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != entity.MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	log.Println("BookHandler.updateBook() - received req", r.Body)

	var book entity.Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Println("BookHandler.updateBook() - received decode error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	var updatedBook entity.Book
	if updatedBook, err = BookHandler.bookStore.UpdateBook(book); err != nil {
		log.Println("BookHandler.updateBook() - received error from db", err)
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("book with id %v wasn't found", book.Id), w)
			return
		}
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(updatedBook)
	if err != nil {
		log.Println("BookHandler.updateBook() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
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

	var invoker entity.User
	if invoker, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), BookHandler.authStore); err != nil {
		log.Println("BookHandler.deleteBook() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != entity.MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		errors.HandleError(403, "403 Forbidden", w)
		return
	}
	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("deleteBook() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	if err = BookHandler.bookStore.Remove(id); err != nil {
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("book with id %v wasn't found", id), w)
			return
		}

		log.Println("deleteBook() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("BookHandler.deleteBook() - successfully finished req", id)

	w.WriteHeader(http.StatusNoContent)
}
