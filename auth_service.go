package main

import (
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func hasPermission(userInfo UserInfo, roles []int) bool {
	return true
}

func parseToken(token string) UserInfo {

	return UserInfo{}
}

func hashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println("hashAndSalt() - received error while generating password", err)
		return "", err
	}

	return string(hash), nil
}

type UserInfo struct {
	id   uuid.UUID
	name string
	role int
}
