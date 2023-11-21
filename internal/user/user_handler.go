package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"example/library-service/internal/auth"
	"example/library-service/internal/entity"
	"example/library-service/internal/errors"
	"example/library-service/internal/utils"

	"github.com/google/uuid"
)

type UserHandler struct {
	userStore *UserStore
	authStore *auth.AuthStore
}

func NewUserHandler(db *sql.DB, authStore *auth.AuthStore) *UserHandler {
	store := NewUserStore(db)
	return &UserHandler{store, authStore}
}

func (userHandler *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && utils.UserRe.Match([]byte(r.URL.Path)):
		userHandler.getUsers(w, r)
		return
	case r.Method == http.MethodGet && utils.UserReWithID.Match([]byte(r.URL.Path)):
		userHandler.getUser(w, r)
		return
	case r.Method == http.MethodPut && utils.UserRe.Match([]byte(r.URL.Path)):
		userHandler.updateUser(w, r)
		return
	case r.Method == http.MethodDelete && utils.UserReWithID.Match([]byte(r.URL.Path)):
		userHandler.deleteUser(w, r)
		return
	default:
		errors.HandleError(405, fmt.Sprintf("Method %v not allowed", r.URL.Path), w)
		return
	}
}

func (userHandler *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("UserHandler.getUser() - processing request", r.URL.Path)

	if err = auth.ValidateToken(r.Header.Get("Authorization"), userHandler.authStore); err != nil {
		log.Println("UserHandler.getUser() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("UserHandler.getUser() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	var User entity.User
	if User, err = userHandler.userStore.GetUser(id); err != nil {
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("user with id %v wasn't found", id), w)
			return
		}
		log.Println("UserHandler.getUser() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(User)
	if err != nil {
		log.Println("UserHandler.getUser() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.getUser() - successfully finished req", User)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (userHandler *UserHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	var err error

	values := r.URL.Query()

	queryMap := utils.ToMap(values)
	log.Println("UserHandler.getUsers() - received req", queryMap)

	if !utils.ValidParams("user", queryMap) {
		errors.HandleError(400, "Invalid request params", w)
		return
	}

	var invoker entity.User
	if invoker, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), userHandler.authStore); err != nil {
		log.Println("UserHandler.getUsers() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != entity.ADMIN {
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	var Users []entity.User
	if Users, err = userHandler.userStore.GetUsers(queryMap); err != nil {
		log.Println("UserHandler.getUsers() - received error from db", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(Users)
	if err != nil {
		log.Println("UserHandler.getUsers() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.getUsers() - successfully finished req", Users)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (userHandler *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	var err error
	var invoker entity.User
	if invoker, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), userHandler.authStore); err != nil {
		log.Println("UserHandler.updateUser() - invalid token: ", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != entity.ADMIN {
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	var user entity.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("UserHandler.updateUser() - received decode error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.updateUser() - received req", user)

	var updatedUser entity.User
	if updatedUser, err = userHandler.userStore.UpdateUser(user); err != nil {
		log.Println("UserHandler.updateUser() - received error from db", err)
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("user with id %v wasn't found", user.Id), w)
			return
		}

		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(updatedUser)
	if err != nil {
		log.Println("UserHandler.updateUser() - received error while marshaling", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.updateUser() - successfully finished req", updatedUser)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")

}

func (userHandler *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("deleteUser() - processing request", r.URL.Path)
	var invoker entity.User
	if invoker, err = auth.ValidateTokenAndGetUser(r.Header.Get("Authorization"), userHandler.authStore); err != nil {
		log.Println("UserHandler.updateUser() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if invoker.Role != entity.ADMIN {
		errors.HandleError(403, "403 Forbidden", w)
		return
	}

	if err = auth.ValidateToken(r.Header.Get("Authorization"), userHandler.authStore); err != nil {
		log.Println("UserHandler.deleteUser() - invalid token", err)
		errors.HandleError(401, err.Error(), w)
		return
	}

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("deleteUser() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	if err = userHandler.userStore.DeleteUser(id); err != nil {
		log.Println("deleteUser() - received error from db", err)
		if err == sql.ErrNoRows {
			errors.HandleError(404, fmt.Sprintf("user with id %v wasn't found", id), w)
			return
		}

		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.deleteUser() - successfully finished req", id)

	w.WriteHeader(http.StatusNoContent)
}
