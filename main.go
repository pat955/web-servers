package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

func main() {
	deleteDB("./database.json")
	db, err := createDB("./database.json")
	if err != nil {
		panic(err)
	}
	db.createChirp("something something")
	fmt.Println(db.loadDB())
	// Use the http.NewServeMux() function to create an empty servemux.
	const root = "."
	const port = "8080"
	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()

	defaultHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	mux.Handle("/app/*", middlewareLog(defaultHandler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCount)
	mux.HandleFunc("GET /api/healthz", handlerStatus)
	mux.HandleFunc("POST /api/chirps", handlerChirp)
	mux.HandleFunc("/api/reset", apiCfg.handlerResetCount)
	corsMux := middlewareCors(mux)

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
	type post struct {
		Body string `json:"body"`
	}
	type cleanedPost struct {
		Body string `json:"cleaned_body"`
	}

	newPOST := post{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newPOST)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	if len(newPOST.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	censored := censor(newPOST.Body)
	respondWithJSON(w, 200, cleanedPost{Body: censored})
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
