package author

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

type AuthorHandler struct {
	authorStore *AuthorStore
	authStore   *auth.AuthStore
}

func NewAuthorHandler(db *sql.DB, authStore *auth.AuthStore) *AuthorHandler {
	store := NewAuthorStore(db)
	return &AuthorHandler{store, authStore}
}

func (authorHandler *AuthorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && utils.AuthorRe.Match([]byte(r.URL.Path)):
		authorHandler.getAuthors(w, r)
		return
	case r.Method == http.MethodGet && utils.AuthorReWithID.Match([]byte(r.URL.Path)):
		authorHandler.getAuthor(w, r)
		return
	case r.Method == http.MethodPost && utils.AuthorRe.Match([]byte(r.URL.Path)):
		authorHandler.createAuthor(w, r)
		return
	case r.Method == http.MethodPut && utils.AuthorRe.Match([]byte(r.URL.Path)):
		authorHandler.updateAuthor(w, r)
		return
	case r.Method == http.MethodDelete && utils.AuthorReWithID.Match([]byte(r.URL.Path)):
		authorHandler.deleteAuthor(w, r)
		return
	default:
		errors.HandleError(405, fmt.Sprintf("Method %v not allowed", r.URL.Path), w)
		return
	}
}

func (AuthorHandler *AuthorHandler) getAuthor(w http.ResponseWriter, r *http.Request) {
	var err error
	var id uuid.UUID
	strs := strings.Split(r.URL.Path, "/")

	log.Println("AuthorHandler.getAuthor() - processing request", r.URL.Path)

	if err = auth.ValidateToken(r.Header.Get("Authorization"), AuthorHandler.authStore); err != nil {
		log.Println("AuthorHandler.getAuthor() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("AuthorHandler.getAuthor() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	var Author entity.Author
	if Author, err = AuthorHandler.authorStore.GetAuthor(id); err != nil {
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("author with id %v wasn't found", id), w)
			return
		}
		log.Println("AuthorHandler.getAuthor() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(Author)
	if err != nil {
		log.Println("AuthorHandler.getAuthor() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthorHandler.getAuthor() - successfully finished req", Author)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (AuthorHandler *AuthorHandler) getAuthors(w http.ResponseWriter, r *http.Request) {
	var err error
	values := r.URL.Query()

	queryMap := utils.ToMap(values)
	log.Println("AuthorHandler.getAuthors() - received req", queryMap)

	if err = auth.ValidateToken(r.Header.Get("Authorization"), AuthorHandler.authStore); err != nil {
		log.Println("AuthorHandler.getAuthors() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if !utils.ValidParams("author", queryMap) {
		log.Println("AuthorHandler.getAuthors() - received invalid params!", queryMap)
		errors.HandleError(400, "Invalid request params", w)
		return
	}

	var Authors []entity.Author
	if Authors, err = AuthorHandler.authorStore.GetAuthors(queryMap); err != nil {
		log.Println("AuthorHandler.getAuthors() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(Authors)
	if err != nil {
		log.Println("AuthorHandler.getAuthors() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthorHandler.getAuthors() - successfully finished req", Authors)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (AuthorHandler *AuthorHandler) createAuthor(w http.ResponseWriter, r *http.Request) {
	var err error
	var user entity.User
	if user, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), AuthorHandler.authStore); err != nil {
		log.Println("AuthorHandler.createAuthor() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if user.Role != entity.MODERATOR {
		log.Println("AuthorHandler.createAuthor() - user doesn't have permission to this resource")
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	var author entity.Author

	if err = json.NewDecoder(r.Body).Decode(&author); err != nil {
		log.Println("AuthorHandler.createAuthor() - received decode error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthorHandler.createAuthor() - received req", author)

	var savedAuthor entity.Author
	if savedAuthor, err = AuthorHandler.authorStore.CreateAuthor(author); err != nil {
		log.Println("AuthorHandler.createAuthor() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(savedAuthor)
	if err != nil {
		log.Println("AuthorHandler.createAuthor() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthorHandler.createAuthor() - successfully finished req", savedAuthor)
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (AuthorHandler *AuthorHandler) updateAuthor(w http.ResponseWriter, r *http.Request) {
	var err error
	var user entity.User
	if user, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), AuthorHandler.authStore); err != nil {
		log.Println("AuthorHandler.updateAuthor() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if user.Role != entity.MODERATOR {
		log.Println("AuthorHandler.updateAuthor() - user doesn't have permission to this resource")
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	var author entity.Author

	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		log.Println("AuthorHandler.updateAuthor() - received decode error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthorHandler.updateAuthor() - received req", author)

	var updatedAuthor entity.Author
	if updatedAuthor, err = AuthorHandler.authorStore.UpdateAuthor(author); err != nil {
		log.Println("AuthorHandler.updateAuthor() - received error from db", err)
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("author with id %v wasn't found", author.Id), w)
			return
		}

		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(updatedAuthor)
	if err != nil {
		log.Println("AuthorHandler.updateAuthor() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
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
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	var user entity.User
	if user, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), AuthorHandler.authStore); err != nil {
		log.Println("AuthorHandler.deleteAuthor() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if user.Role != entity.MODERATOR {
		log.Println("AuthorHandler.deleteAuthor() - user doesn't have permission to this resource")
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	if err = AuthorHandler.authorStore.DeleteAuthor(id); err != nil {
		log.Println("deleteAuthor() - received error from db", err)
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("author with id %v wasn't found", id), w)
			return
		}

		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthorHandler.deleteAuthor() - successfully finished req", id)

	w.WriteHeader(http.StatusNoContent)
}
