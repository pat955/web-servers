package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pat955/chirpy/internal/auth"
	"github.com/pat955/chirpy/internal/my_db"
	"golang.org/x/crypto/bcrypt"
)

func handlerAuth(w http.ResponseWriter, req *http.Request) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		respondWithError(w, 500, "JWT secret not set")
		return
	}

	db := my_db.CreateDB(DBPATH)

	status, tokenString := auth.GetAuthFromRequest(req)
	if status > 201 {
		respondWithError(w, status, tokenString)
		return
	}
	token, err := auth.GetToken(tokenString, jwtSecret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized response, "+err.Error())
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		respondWithError(w, 401, "Invalid token")
		return
	}

	email := claims.Subject
	if email == "" {
		respondWithError(w, 401, "Token subject is missing")
		return
	}
	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		panic(err)
	}
	foundUser, found := db.GetUser(id)
	if !found {
		respondWithError(w, 404, "User not found")
		return
	}
	var user my_db.User
	my_db.DecodeForm(req, &user)

	foundUser.Email = user.Email
	foundUser.Password = user.Password
	db.UpdateUser(foundUser)

	respondWithJSON(w, 200, foundUser.UserToPublic())
}

func handlerLogin(w http.ResponseWriter, req *http.Request) {
	db := my_db.CreateDB(DBPATH)

	var user my_db.User
	my_db.DecodeForm(req, &user)

	var id int
	for _, u := range db.GetUsers() {
		if u.Email == user.Email {
			id = u.ID
		}
	}
	if id == 0 {
		respondWithError(w, 404, "user not found")
		return
	}
	foundUser, found := db.GetUser(id)
	if !found {
		respondWithError(w, 404, "user not found")
		return
	}
	foundUser.ExpiresInSeconds = user.ExpiresInSeconds
	err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))
	if err != nil {
		respondWithError(w, 401, "wrong password")
		return
	}
	db.UpdateUser(foundUser)
	respondWithJSON(w, 200, foundUser.UserLoginResponse())
}
