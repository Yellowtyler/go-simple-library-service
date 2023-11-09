package main

import "net/http"

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
	case r.Method == http.MethodPost && r.URL.Path == LoginPath:
		authHandler.login(w, r)
	case r.Method == http.MethodPost && r.URL.Path == LogoutPath:
		authHandler.logout(w, r)
	}
}

func (authHandler *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	// parse request
	// validate name and mail
	// create user
}

func (authHandler *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	return
}

func (AuthHandler *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {

}

type ReqisterRequest struct {
	name     string
	mail     string
	password string
	role     int
}

type LoginRequest struct {
	name     string
	password string
}
