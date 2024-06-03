package main

import (
	"flag"
	"log"
	"os"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pat955/chirpy/internal/my_db"
)

const DBPATH string = "./database.json"

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")

	// debug flag, deletes the db if $ ./out --debug
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		my_db.DeleteDB(DBPATH)
	}

	const root = "../public"
	const port = "8080"
	apiCfg := apiConfig{
		DB:             my_db.CreateDB(DBPATH),
		fileserverHits: 0,
		JWTSecret:      jwtSecret,
	}

	router := mux.NewRouter()
	defaultHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	router.Handle("/app/*", middlewareLog(defaultHandler))
	router.HandleFunc("/admin/metrics", apiCfg.handlerCount).Methods("GET")
	router.HandleFunc("/api/healthz", handlerStatus).Methods("GET")
	router.HandleFunc("/api/chirps", handlerAddChirp).Methods("POST")
	router.HandleFunc("/api/chirps", handlerGetChirps).Methods("GET")
	router.HandleFunc("/api/chirps/{chirpID}", handlerAddChirpId).Methods("GET")
	router.HandleFunc("/api/users", handlerAddUser).Methods("POST")
	router.HandleFunc("/api/users", handlerAuth).Methods("PUT")
	router.HandleFunc("/api/login", handlerLogin).Methods("POST")
	router.HandleFunc("/api/reset", apiCfg.handlerResetCount)
	corsMux := middlewareLog(middlewareCors(router))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	srv.ListenAndServe()
}

// decodes json into your provided struct. Using this to avoid making a massive all encompassing struct
