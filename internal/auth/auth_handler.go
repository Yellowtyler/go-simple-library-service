package auth

import (
	"encoding/json"
	"example/library-service/internal/entity"
	"example/library-service/internal/errors"
	"example/library-service/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	S *AuthStore
}

func NewAuthHandler(s *AuthStore) *AuthHandler {
	return &AuthHandler{s}
}

func (authHandler *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == utils.RegisterPath:
		authHandler.register(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == utils.LoginPath:
		authHandler.login(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == utils.LogoutPath:
		authHandler.logout(w, r)
		return
	default:
		errors.HandleError(405, fmt.Sprintf("Method %v not allowed", r.URL.Path), w)
		return
	}
}

func (authHandler *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req ReqisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("AuthHandler.register() - error while decoding", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthHandler.register() - started to process", req)

	var exists bool
	var err error
	if exists, err = authHandler.S.ExistsWithNameOrMail(req.Name, req.Mail); err != nil {
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	if exists {
		errors.HandleError(400, fmt.Sprintf("user %v already exists!", req.Name), w)
		return
	}

	req.Password, err = HashAndSalt([]byte(req.Password))
	if err != nil {
		log.Println("AuthHandler.register() - error while hashing password", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	if err := authHandler.S.CreateUser(req); err != nil {
		errors.HandleError(500, "Internal Server Error", w)
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
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	log.Println("AuthHandler.login() - started to process", req.Name)

	var user entity.User
	var err error
	if user, err = authHandler.S.GetUserByName(req.Name); err != nil {
		errors.HandleError(401, "wrong username", w)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		errors.HandleError(401, "wrong password", w)
		return
	}

	var token string

	if token, err = GenerateToken(user.Id, user.Role); err != nil {
		log.Println("AuthHandler.login() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
		return
	}

	user.Token = token
	if updateErr := authHandler.S.UpdateToken(user); updateErr != nil {
		log.Println("AuthHandler.login() - received error", err)
		errors.HandleError(500, "Internal Server Error", w)
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
		errors.HandleError(401, "Authorization header wasn't provided", w)
		return
	}

	id, _, err := ParseToken(token)
	if err != nil && err.Error() != "token is expired!" {
		errors.HandleError(500, "Internal Server Error", w)
		return
	}
	if err := authHandler.S.DeleteToken(id); err != nil {
		errors.HandleError(500, "Internal Server Error", w)
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
