package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("VeryVeryVeryVeryVeryVeryBigSecretKey")

const expirationTime int64 = 100

func hasPermission(userId uuid.UUID, role int, roles []int) bool {
	return true
}

func ParseToken(tokenString string) (uuid.UUID, int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Println("AuthService.ParseToken() - unexpected signing method")
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secretKey, nil
	})

	if err != nil {
		log.Println("AuthService.ParseToken() - received error ", err)
		return uuid.Nil, 0, err
	}

	var id string
	var role int
	var expiredAt int64
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		log.Println("AuthService.ParseToken() - claims", claims)
		id = claims["id"].(string)
		role = int(claims["role"].(float64))
		expiredAt = int64(claims["expired_at"].(float64))
	} else {
		log.Println(err)
		return uuid.Nil, role, err
	}

	if time.Now().Unix() >= expiredAt {
		log.Println("AuthService.ParseToken() - token is expired!")
		return uuid.Nil, role, fmt.Errorf("token is expired!")
	}

	uid := uuid.MustParse(id)

	return uid, role, nil
}

func GenerateToken(id uuid.UUID, role int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         id,
		"role":       role,
		"expired_at": time.Now().Unix() + expirationTime,
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
		log.Println("HashAndSalt() - received error while generating password", err)
		return "", err
	}

	return string(hash), nil
}
