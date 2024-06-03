package main

import (
	"net/http"

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
