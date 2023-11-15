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

type UserHandler struct {
	s *UserStore
}

func NewUserHandler(store *UserStore) *UserHandler {
	return &UserHandler{store}
}

func (userHandler *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && UserRe.Match([]byte(r.URL.Path)):
		userHandler.getUsers(w, r)
		return
	case r.Method == http.MethodGet && UserReWithID.Match([]byte(r.URL.Path)):
		userHandler.getUser(w, r)
		return
	case r.Method == http.MethodPut && UserRe.Match([]byte(r.URL.Path)):
		userHandler.updateUser(w, r)
		return
	case r.Method == http.MethodDelete && UserReWithID.Match([]byte(r.URL.Path)):
		userHandler.deleteUser(w, r)
		return
	default:
		HandleError(405, fmt.Sprintf("Method %v not allowed", r.URL.Path), w)
		return
	}
}

func (userHandler *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID
	var err error
	strs := strings.Split(r.URL.Path, "/")

	log.Println("UserHandler.getUser() - processing request", r.URL.Path)

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("UserHandler.getUser() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	var User User
	if User, err = userHandler.s.GetUser(id); err != nil {
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("user with id %v wasn't found", id), w)
			return
		}
		log.Println("UserHandler.getUser() - received error from db", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(User)
	if err != nil {
		log.Println("UserHandler.getUser() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.getUser() - successfully finished req", User)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (userHandler *UserHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	queryMap := ToMap(values)
	log.Println("UserHandler.getUsers() - received req", queryMap)

	if !ValidParams("User", queryMap) {
		log.Println("UserHandler.getUsers() - received invalid params!", queryMap)

		HandleError(500, "Internal Server Error", w)
		return
	}

	var Users []User
	var err error
	if Users, err = userHandler.s.GetUsers(queryMap); err != nil {
		log.Println("UserHandler.getUsers() - received error from db", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(Users)
	if err != nil {
		log.Println("UserHandler.getUsers() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.getUsers() - successfully finished req", Users)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Header().Set("Content-Type", "application/json")
}

func (userHandler *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("UserHandler.updateUser() - received req", r.Body)

	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("UserHandler.updateUser() - received decode error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	var updatedUser User
	var err error
	if updatedUser, err = userHandler.s.UpdateUser(user); err != nil {
		log.Println("UserHandler.updateUser() - received error from db", err)
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("user with id %v wasn't found", user.Id), w)
			return
		}

		HandleError(500, "Internal Server Error", w)
		return
	}

	jsonBytes, err := json.Marshal(updatedUser)
	if err != nil {
		log.Println("UserHandler.updateUser() - received error while marshaling", err)
		HandleError(500, "Internal Server Error", w)
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

	if id, err = uuid.Parse(strs[len(strs)-1]); err != nil {
		log.Println("deleteUser() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	if err = userHandler.s.DeleteUser(id); err != nil {
		log.Println("deleteUser() - received error from db", err)
		if err == sql.ErrNoRows {
			HandleError(404, fmt.Sprintf("user with id %v wasn't found", id), w)
			return
		}

		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("UserHandler.deleteUser() - successfully finished req", id)

	w.WriteHeader(http.StatusNoContent)
}
