package main

import (
	"fmt"
	"net/http"

	"github.com/Yellowtyler/handlers"
)

func main() {
	fmt.Printf("Hi!")
	server := http.NewServeMux()
	server.Handle("/", &homeHandler{})
	server.Handle("/books", &handlers.BookHandler{})
	server.Handle("/books/", &handlers.BookHandler{})
	http.ListenAndServe(":8080", server)
}

type homeHandler struct{}

func (handler *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Test"))
}
