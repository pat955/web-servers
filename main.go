package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
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
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
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

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	if statusCode > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, statusCode, errorResponse{
		Error: msg,
	})
}
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type post struct {
		Body string
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	newPOST := post{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newPOST)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		fmt.Println("POST FAILED", err)
		return
	}
	if len(newPOST.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	respondWithJSON(w, 200, returnVals{Valid: true})
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
