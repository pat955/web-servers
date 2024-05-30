package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

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
	router := mux.NewRouter()

	defaultHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	router.Handle("/app/*", middlewareLog(defaultHandler))
	router.HandleFunc("/admin/metrics", apiCfg.handlerCount).Methods("GET")
	router.HandleFunc("/api/healthz", handlerStatus).Methods("GET")
	router.HandleFunc("/api/chirps", handlerChirp).Methods("POST")
	router.HandleFunc("/api/chirps", handlerGetChirps).Methods("GET")
	router.HandleFunc("/api/reset", apiCfg.handlerResetCount).Methods("GET")
	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	srv.ListenAndServe()
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func handlerStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerChirp(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	type post struct {
		Body string `json:"body"`
	}

	newPOST := post{}
	json.NewDecoder(req.Body).Decode(&newPOST)
	if len(newPOST.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	newChirp, err := db.createChirp(newPOST.Body)
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

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func censor(s string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	return re.ReplaceAllString(s, "****")
}
