package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type AuthorHandler struct {
	s *AuthorStore
}

func NewAuthorHandler(db *sql.DB) *AuthorHandler {
	store := NewAuthorStore(db)
	return &AuthorHandler{store}
}

func (authorHandler *AuthorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && AuthorRe.Match([]byte(r.URL.Path)):
		authorHandler.getAuthors(w, r)
		return
	case r.Method == http.MethodGet && AuthorReWithID.Match([]byte(r.URL.Path)):
		authorHandler.getAuthor(w, r)
		return
	case r.Method == http.MethodPost && AuthorRe.Match([]byte(r.URL.Path)):
		authorHandler.createAuthor(w, r)
		return
	case r.Method == http.MethodPut && AuthorRe.Match([]byte(r.URL.Path)):
		authorHandler.updateAuthor(w, r)
		return
	case r.Method == http.MethodDelete && AuthorReWithID.Match([]byte(r.URL.Path)):
		authorHandler.deleteAuthor(w, r)
		return
	default:
		MethodNotAllowedHandler(w, r)
		return
	}
}

func (AuthorHandler *AuthorHandler) getAuthor(w http.ResponseWriter, r *http.Request) {
	var err error

	if err = ValidateToken(r.Header.Get("Authorization"), (*UserStore)(AuthorHandler.s)); err != nil {
		if err == sql.ErrNoRows {
			log.Println("AuthorHandler.getAuthor() - user wasn't found")
			UnauthorizedHandler(w, r, "user wasn't found")
			return
		}
		log.Println("AuthorHandler.getAuthor() - invalid token", err)
		UnauthorizedHandler(w, r, err.Error())
		return
	}

	var id uuid.UUID
	strs := strings.Split(r.URL.Path, "/")

	log.Println("AuthorHandler.getAuthor() - processing request", r.URL.Path)

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("AuthorHandler.getAuthor() - received error", err)
		InternalServerErrorHandler(w, r)
		return
	}

	var Author Author
	if Author, err = AuthorHandler.s.GetAuthor(id); err != nil {
		if err == sql.ErrNoRows {
			NotFoundHandler(w, r)
			return
		}
		log.Println("AuthorHandler.getAuthor() - received error from db", err)
		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(Author)
	if err != nil {
		log.Println("AuthorHandler.getAuthor() - received error while marshaling", err)
		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.getAuthor() - successfully finished req", Author)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (AuthorHandler *AuthorHandler) getAuthors(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = ValidateToken(r.Header.Get("Authorization"), (*UserStore)(AuthorHandler.s)); err != nil {
		if err == sql.ErrNoRows {
			log.Println("AuthorHandler.getAuthors() - user wasn't found")
			UnauthorizedHandler(w, r, "user wasn't found")
			return
		}
		log.Println("AuthorHandler.getAuthors() - invalid token", err)
		UnauthorizedHandler(w, r, err.Error())
		return
	}

	values := r.URL.Query()

	queryMap := ToMap(values)
	log.Println("AuthorHandler.getAuthors() - received req", queryMap)

	if !ValidParams("author", queryMap) {
		log.Println("AuthorHandler.getAuthors() - received invalid params!", queryMap)

		InternalServerErrorHandler(w, r)
		return
	}

	var Authors []Author
	if Authors, err = AuthorHandler.s.GetAuthors(queryMap); err != nil {
		log.Println("AuthorHandler.getAuthors() - received error from db", err)
		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(Authors)
	if err != nil {
		log.Println("AuthorHandler.getAuthors() - received error while marshaling", err)
		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.getAuthors() - successfully finished req", Authors)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (AuthorHandler *AuthorHandler) createAuthor(w http.ResponseWriter, r *http.Request) {
	var err error
	var user User
	if user, err = ValidateTokenAndGetUser(r.Header.Get("Authorization"), (*UserStore)(AuthorHandler.s)); err != nil {
		if err == sql.ErrNoRows {
			log.Println("AuthorHandler.createAuthor() - user wasn't found")
			UnauthorizedHandler(w, r, "user wasn't found")
			return
		}
		log.Println("AuthorHandler.createAuthor() - invalid token", err)
		UnauthorizedHandler(w, r, err.Error())
		return
	}

	if user.Role != MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		ForbiddenHandler(w, r)
		return
	}
	var author Author

	if err = json.NewDecoder(r.Body).Decode(&author); err != nil {
		log.Println("AuthorHandler.createAuthor() - received decode error", err)
		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.createAuthor() - received req", author)

	var savedAuthor Author
	if savedAuthor, err = AuthorHandler.s.CreateAuthor(author); err != nil {
		log.Println("AuthorHandler.createAuthor() - received error from db", err)
		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(savedAuthor)
	if err != nil {
		log.Println("AuthorHandler.createAuthor() - received error while marshaling", err)
		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.createAuthor() - successfully finished req", savedAuthor)
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (AuthorHandler *AuthorHandler) updateAuthor(w http.ResponseWriter, r *http.Request) {
	var err error
	var user User
	if user, err = ValidateTokenAndGetUser(r.Header.Get("Authorization"), (*UserStore)(AuthorHandler.s)); err != nil {
		if err == sql.ErrNoRows {
			log.Println("AuthorHandler.updateAuthor() - user wasn't found")
			UnauthorizedHandler(w, r, "user wasn't found")
			return
		}
		log.Println("AuthorHandler.updateAuthor() - invalid token", err)
		UnauthorizedHandler(w, r, err.Error())
		return
	}

	if user.Role != MODERATOR {
		log.Println("AuthorHandler.updateAuthor() - user doesn't have permission to this resource")
		ForbiddenHandler(w, r)
		return
	}

	var author Author

	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		log.Println("AuthorHandler.updateAuthor() - received decode error", err)
		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.updateAuthor() - received req", author)

	var updatedAuthor Author
	if updatedAuthor, err = AuthorHandler.s.UpdateAuthor(author); err != nil {
		log.Println("AuthorHandler.updateAuthor() - received error from db", err)
		if err == sql.ErrNoRows {
			NotFoundHandler(w, r)
			return
		}

		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(updatedAuthor)
	if err != nil {
		log.Println("AuthorHandler.updateAuthor() - received error while marshaling", err)
		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.updateAuthor() - successfully finished req", updatedAuthor)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (AuthorHandler *AuthorHandler) deleteAuthor(w http.ResponseWriter, r *http.Request) {
	var err error
	var id uuid.UUID
	strs := strings.Split(r.URL.Path, "/")

	log.Println("deleteAuthor() - processing request", r.URL.Path)

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("deleteAuthor() - received error", err)
		InternalServerErrorHandler(w, r)
		return
	}

	var user User
	if user, err = ValidateTokenAndGetUser(r.Header.Get("Authorization"), (*UserStore)(AuthorHandler.s)); err != nil {
		if err == sql.ErrNoRows {
			log.Println("AuthorHandler.deleteAuthor() - user wasn't found")
			UnauthorizedHandler(w, r, "user wasn't found")
			return
		}
		log.Println("AuthorHandler.deleteAuthor() - invalid token", err)
		UnauthorizedHandler(w, r, err.Error())
		return
	}

	if user.Role != MODERATOR {
		log.Println("AuthorHandler.deleteAuthor() - user doesn't have permission to this resource")
		ForbiddenHandler(w, r)
		return
	}

	if err = AuthorHandler.s.DeleteAuthor(id); err != nil {
		log.Println("deleteAuthor() - received error from db", err)
		if err == sql.ErrNoRows {
			NotFoundHandler(w, r)
			return
		}

		InternalServerErrorHandler(w, r)
		return
	}

	log.Println("AuthorHandler.deleteAuthor() - successfully finished req", id)

	w.WriteHeader(http.StatusNoContent)
}
