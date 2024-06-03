package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pat955/chirpy/internal/my_db"
)

var CHIRPID int = 1

func handlerAddChirp(w http.ResponseWriter, req *http.Request) {
	db := my_db.CreateDB(DBPATH)

	var chirp my_db.Chirp
	my_db.DecodeForm(req, &chirp)
	chirp.Body = my_db.Censor(chirp.Body)

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	chirp.ID = CHIRPID
	db.AddChirp(chirp)
	CHIRPID++
	respondWithJSON(w, 201, chirp)
}

func handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	respondWithJSON(w, 200, my_db.CreateDB(DBPATH).GetChirps())
}

func handlerAddChirpId(w http.ResponseWriter, req *http.Request) {
	chirpID, ok := mux.Vars(req)["chirpID"]
	if !ok {
		respondWithError(w, 400, "id is missing in parameters")
		return
	}
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	db := my_db.CreateDB(DBPATH)

	chirp, found := db.GetChirpMap()[id]
	if !found {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, chirp)
}
