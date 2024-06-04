package main

import (
	"net/http"
	"os"

	"github.com/pat955/chirpy/internal/auth"
	"github.com/pat955/chirpy/internal/my_db"
)

var USERID int = 1

func handlerAddUser(w http.ResponseWriter, req *http.Request) {
	db := my_db.CreateDB(DBPATH)

	var user my_db.User
	my_db.DecodeForm(req, &user)
	_, found := db.GetUser(user.ID)
	if found {
		respondWithError(w, 409, "user already exists")
		return
	}
	user.ID = USERID
	db.AddUser(user)
	USERID++
	respondWithJSON(w, 201, user.UserToPublic())
}

func handlerRefresh(w http.ResponseWriter, req *http.Request) {
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
	data := db.GetUsers()
	for _, u := range data {
		if u.AccessToken == tokenString {
			respondWithJSON(w, 201, RefreshResponse{Token: u.AccessToken})
			return
		}
	}
	respondWithError(w, 404, "access token not found")
}

func handlerRevoke(w http.ResponseWriter, req *http.Request) {
	status, tokenString := auth.GetAuthFromRequest(req)
	if status > 204 {
		respondWithError(w, status, tokenString)
		return
	}
	db := my_db.CreateDB(DBPATH)
	db.Revoke(tokenString)

	respondWithJSON(w, 204, nil)
}

type RefreshResponse struct {
	Token string `json:"token"`
}
