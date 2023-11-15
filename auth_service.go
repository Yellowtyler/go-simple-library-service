package main

import (
	"fmt"
	"log"
	"strings"
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

func ValidateTokenAndGetUser(authHeader string, store *UserStore) (user User, err error) {
	if authHeader == "" {
		return user, fmt.Errorf("empty Authorization header")
	}

	vals := strings.Split(authHeader, " ")
	if len(vals) < 2 {
		return user, fmt.Errorf("wrong header value")
	}

	token := vals[1]

	var id uuid.UUID
	var role int

	if id, role, err = ParseToken(token); err != nil {
		return user, err
	}

	if user, err = store.GetUserByIdAndRole(id, role); err != nil {
		return user, fmt.Errorf("invalid token")
	}

	return user, nil
}

func ValidateToken(authHeader string, store *UserStore) error {
	if authHeader == "" {
		return fmt.Errorf("empty Authorization header")
	}

	vals := strings.Split(authHeader, " ")
	if len(vals) < 2 {
		return fmt.Errorf("wrong header value")
	}

	token := vals[1]

	var id uuid.UUID
	var role int
	var err error
	if id, role, err = ParseToken(token); err != nil {
		return err
	}

	if _, err = store.GetUserByIdAndRole(id, role); err != nil {
		return fmt.Errorf("invalid token")
	}

	return nil
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
		log.Println("AuthService.GenerateToken() received error while signing", err)
		return "", err
	}

	return tokenString, nil
}

func HashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println("AuthService.HashAndSalt() - received error while generating password", err)
		return "", err
	}

	return string(hash), nil
}
