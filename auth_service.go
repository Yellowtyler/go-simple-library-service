package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const secretKey string = "VeryVeryVeryVeryVeryVeryBigSecretKey"
const expirationTime int64 = 100

func hasPermission(userId uuid.UUID, role int, roles []int) bool {
	return true
}

func ParseToken(tokenString string) (uuid.UUID, int, error) {
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
	var issuedAt int64
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		id = claims["id"].(uuid.UUID)
		role = claims["role"].(int)
		issuedAt = claims["issued_at"].(int64)
	} else {
		fmt.Println(err)
		return id, role, err
	}

	if time.Now().Unix()-issuedAt >= expirationTime {
		return id, role, fmt.Errorf("token is expired!")
	}

	return id, role, nil
}

func GenerateToken(id uuid.UUID, role int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":        id,
		"role":      role,
		"issued_at": time.Now().Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		log.Println("GenerateToken() received error while signing", err)
		return "", err
	}

	return tokenString, nil
}

func HashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println("hashAndSalt() - received error while generating password", err)
		return "", err
	}

	return string(hash), nil
}
