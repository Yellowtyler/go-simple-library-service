package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("started app")
	db := Connect()
	db.Ping()

	bookHandler := NewBookHandler(db)
	authorHandler := NewAuthorHandler(db)
	server := http.NewServeMux()
	server.Handle("/", &homeHandler{})
	server.Handle("/books", bookHandler)
	server.Handle("/books/", bookHandler)
	server.Handle("/authors", authorHandler)
	server.Handle("/authors/", authorHandler)
	http.ListenAndServe(":8080", server)
}

type homeHandler struct{}

func (handler *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Test"))
}
