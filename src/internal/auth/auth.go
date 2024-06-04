package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Login struct {
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func JWTNotSetCheck() error {
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret == "" {
		return errors.New("jwt not set")
	}
	return nil
}
func GetAuthFromRequest(req *http.Request) (int, string) {
	auth := req.Header.Get("Authorization")
	if auth == "Bearer: " || auth == "" {
		return 401, "Authorization header missing"
	}

	tokenString := strings.Split(auth, "Bearer ")
	if len(tokenString) != 2 {
		return 401, "Invalid Authorization header format"
	}
	return 200, tokenString[1]
}

func GetToken(tokenString, jwtSecret string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	return token, err
}
