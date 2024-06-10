package main

import (
	"net/http"
	"os"

	"github.com/pat955/chirpy/internal/auth"
	"github.com/pat955/chirpy/internal/my_db"
)

type Event struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func handlerUpgraded(w http.ResponseWriter, req *http.Request) {
	db := my_db.CreateDB(DBPATH)
	var e Event
	my_db.DecodeForm(req, &e)

	status, authString := auth.GetAuthFromRequest(req)
	if status > 204 {
		respondWithError(w, status, authString)
		return
	}
	if !verifyPolkaApiKey(authString) {
		respondWithError(w, 401, "Invalid API key")
		return
	}
	if e.Event != "user.upgraded" {
		respondWithJSON(w, 204, nil)
		return
	}
	u, found := db.GetUser(e.Data.UserID)
	if !found {
		respondWithError(w, 404, "userid does not exist")
		return
	}
	u.IsChirpyRed = true
	db.UpdateUser(u)
	respondWithJSON(w, 204, nil)
}

func verifyPolkaApiKey(key string) bool {
	POLKA_KEY := os.Getenv("POLKA_KEY")
	return key == POLKA_KEY
}
