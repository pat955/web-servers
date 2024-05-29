package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
)

const DBPATH string = "./database.json"

func main() {
	deleteDB(DBPATH)

	// Use the http.NewServeMux() function to create an empty servemux.
	const root = "."
	const port = "8080"
	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	r := mux.NewRouter()

	defaultHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	r.Handle("/app/*", middlewareLog(defaultHandler))
	r.HandleFunc("/admin/metrics", apiCfg.handlerCount).Methods("GET")
	r.HandleFunc("/api/healthz", handlerStatus).Methods("GET")
	r.HandleFunc("/api/chirps", handlerAddChirp).Methods("POST")
	r.HandleFunc("/api/chirps", handlerGetChirps).Methods("GET")
	r.HandleFunc("/api/chirps/{chirpID}", handlerAddChirpId).Methods("GET")
	r.HandleFunc("/api/users", handlerAddUser).Methods("POST")
	r.HandleFunc("/api/reset", apiCfg.handlerResetCount)
	corsMux := middlewareLog(middlewareCors(r))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	srv.ListenAndServe()
}

func handlerAddChirp(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	chirp := POST{}
	json.NewDecoder(req.Body).Decode(&chirp)

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	newChirp, err := db.createChirp(censor(chirp.Body))
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 201, newChirp)
}

func handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 200, db.getChirps())
}

func handlerAddChirpId(w http.ResponseWriter, req *http.Request) {
	chirpID, ok := mux.Vars(req)["chirpID"]
	if !ok {
		fmt.Println("id is missing in parameters")
	}
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	chirpMap, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	chirp, found := chirpMap.Chrips[id]
	if !found {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, chirp)
}

func handlerAddUser(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	fmt.Println(req.GetBody())
	user := POST{}
	json.NewDecoder(req.Body).Decode(&user)

	newUser, err := db.createUser(user.Body)
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 201, newUser)
}

func censor(s string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	return re.ReplaceAllString(s, "****")
}

type POST struct {
	Body string `json:"body"`
}
