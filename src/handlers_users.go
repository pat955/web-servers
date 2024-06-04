package main

import (
	"net/http"
	"time"

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
	status, refreshToken := auth.GetAuthFromRequest(req)
	if status > 204 {
		respondWithError(w, 204, refreshToken)
	}
	db := my_db.CreateDB(DBPATH)
	token := db.GetRefreshToken(refreshToken)
	if time.Now().Compare(token.ExpiresUTC) == 1 || token.UserID == 0 {
		respondWithError(w, 401, "Expired refresh token, or non existant token")
		return
	}
	u, _ := db.GetUser(token.UserID)
	respondWithJSON(w, 200, RefreshResponse{Token: u.GenerateToken()})

	// or if it's expired, respond with a 401 status code. Otherwise, respond with a 200 code and this shape:

	//	{
	//	    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	//	}
	//
	// The token field should be a newly created access token that expires in 1 hour.
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
