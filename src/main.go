package main

import (
	"flag"
	"fmt"
	"os"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pat955/chirpy/internal/my_db"
)

const DBPATH string = "./database.json"

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		panic(fmt.Sprintf("err loading: %v", err))
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT SECRET NOT SET")
	}
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
	router.HandleFunc("/api/chirps/{chirpID}", handlerDeleteChirp).Methods("DELETE")
	router.HandleFunc("/api/users", handlerAddUser).Methods("POST")
	router.HandleFunc("/api/users", handlerAuth).Methods("PUT")
	router.HandleFunc("/api/refresh", handlerRefresh).Methods("POST")
	router.HandleFunc("/api/revoke", handlerRevoke).Methods("POST")
	router.HandleFunc("/api/login", handlerLogin).Methods("POST")
	router.HandleFunc("/api/reset", apiCfg.handlerResetCount)

	router.HandleFunc("/api/polka/webhooks", handlerUpgraded).Methods("POST")
	corsMux := middlewareLog(middlewareCors(router))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	srv.ListenAndServe()
}
