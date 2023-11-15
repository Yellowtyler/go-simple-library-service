package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	s *UserStore
}

func NewAuthHandler(s *UserStore) *AuthHandler {
	return &AuthHandler{s}
}

func (authHandler *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == RegisterPath:
		authHandler.register(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == LoginPath:
		authHandler.login(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == LogoutPath:
		authHandler.logout(w, r)
		return
	default:
		HandleError(405, fmt.Sprintf("Method %v not allowed", r.URL.Path), w)
		return
	}
}

func (authHandler *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req ReqisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("AuthHandler.register() - error while decoding", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthHandler.register() - started to process", req)

	var exists bool
	var err error
	if exists, err = authHandler.s.ExistsWithNameOrMail(req.Name, req.Mail); err != nil {
		HandleError(500, "Internal Server Error", w)
		return
	}

	if exists {
		HandleError(400, fmt.Sprintf("user %v already exists!", req.Name), w)
		return
	}

	req.Password, err = HashAndSalt([]byte(req.Password))
	if err != nil {
		log.Println("AuthHandler.register() - error while hashing password", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	if err := authHandler.s.CreateUser(req); err != nil {
		HandleError(500, "Internal Server Error", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	log.Println("AuthHandler.register() - finished to process", req)
}

func (authHandler *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("AuthHandler.login() - error while decoding", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthHandler.login() - started to process", req.Name)

	var user User
	var err error
	if user, err = authHandler.s.GetUserByName(req.Name); err != nil {
		HandleError(401, "wrong username", w)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		HandleError(401, "wrong password", w)
		return
	}

	var token string

	if token, err = GenerateToken(user.Id, user.Role); err != nil {
		log.Println("AuthHandler.login() - received error", err)
		HandleError(500, "Internal Server Error", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(token))

	log.Println("AuthHandler.login() - finished to process", req)

}

func (authHandler *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	log.Println("AuthHandler.logout() - started to process")
	token := strings.Split(r.Header.Get("Authorization"), " ")[1]
	if token == "" {
		HandleError(401, "Authorization header wasn't provided", w)
		return
	}

	id, _, err := ParseToken(token)
	if err != nil {
		HandleError(500, "Internal Server Error", w)
		return
	}
	if err := authHandler.s.DeleteToken(id); err != nil {
		HandleError(500, "Internal Server Error", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	log.Println("AuthHandler.logout() - successfully ended to process")
}

type ReqisterRequest struct {
	Name     string `json:"name"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Role     int    `json:"role"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (r ReqisterRequest) String() string {
	return fmt.Sprintf("name: %v, mail: %v", r.Name, r.Mail)
}

func (r LoginRequest) String() string {
	return fmt.Sprintf("name: %v", r.Name)
}
