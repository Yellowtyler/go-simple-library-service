package main

import (
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const secretKey string = "VeryVeryVeryVeryVeryVeryBigSecretKey"

func hasPermission(userId uuid.UUID, role int, roles []int) bool {
	return true
}

func parseToken(tokenString string) (uuid.UUID, int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secretKey, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	var id uuid.UUID
	var role int
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		id = claims["id"].(uuid.UUID)
		role = claims["role"].(int)
	} else {
		fmt.Println(err)
	}

	return id, role, nil
}

func generateToken(id uuid.UUID, role int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   id,
		"role": role,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func hashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println("hashAndSalt() - received error while generating password", err)
		return "", err
	}

	return string(hash), nil
}
