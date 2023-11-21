package main

import (
	"example/library-service/internal/auth"
	"example/library-service/internal/author"
	"example/library-service/internal/book"
	"example/library-service/internal/user"
	"log"
	"net/http"
)

func main() {
	log.Println("main.starting app...")
	db := Connect()
	db.Ping()
	authStore := auth.NewAuthStore(db)
	bookHandler := book.NewBookHandler(db, authStore)
	authorHandler := author.NewAuthorHandler(db, authStore)
	userHandler := user.NewUserHandler(db, authStore)
	authHandler := auth.NewAuthHandler(authStore)
	server := http.NewServeMux()
	server.Handle("/books", bookHandler)
	server.Handle("/books/", bookHandler)
	server.Handle("/authors", authorHandler)
	server.Handle("/authors/", authorHandler)
	server.Handle("/users", userHandler)
	server.Handle("/users/", userHandler)
	server.Handle("/auth/", authHandler)
	http.ListenAndServe(":8080", server)
}
