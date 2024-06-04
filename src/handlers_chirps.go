package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pat955/chirpy/internal/auth"
	"github.com/pat955/chirpy/internal/my_db"
)

var CHIRPID int = 1

func getAuthorId(req *http.Request) (int, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	status, token := auth.GetAuthFromRequest(req)
	if status > 201 {
		return -1, errors.New(token)
	}

	jwt, err := auth.GetToken(token, jwtSecret)
	if err != nil {
		return -1, err
	}
	id, err := jwt.Claims.GetSubject()
	if err != nil {
		return -1, err
	}
	return strconvInt(id), nil
}

func handlerAddChirp(w http.ResponseWriter, req *http.Request) {
	authorID, err := getAuthorId(req)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	db := my_db.CreateDB(DBPATH)

	var chirp my_db.Chirp
	my_db.DecodeForm(req, &chirp)

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	chirp.ID = CHIRPID
	chirp.AuthorID = authorID
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

	chirp, found := db.GetChirp(id)
	if !found {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, chirp)
}

func handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	chirpID, ok := mux.Vars(req)["chirpID"]
	if !ok {
		respondWithError(w, 400, "id is missing in parameters")
		return
	}
	authorID, err := getAuthorId(req)
	if err != nil {
		respondWithError(w, 403, "No authorization")
		return
	}
	db := my_db.CreateDB(DBPATH)
	if err := db.DeleteChirp(strconvInt(chirpID), authorID); err != nil {
		respondWithError(w, 403, err.Error())
		return
	}
	respondWithJSON(w, 204, nil)
}

func strconvInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
