package handlers

import (
	"net/http"
)

type BookHandler struct{}

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
	w.Write([]byte("get book"))
}

func (BookHandler *BookHandler) getBooks(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get books"))

}

func (BookHandler *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create book"))

}

func (BookHandler *BookHandler) updateBook(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("update book"))

}

func (BookHandler *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("delete book"))
}
