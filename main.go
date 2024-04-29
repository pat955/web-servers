package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Use the http.NewServeMux() function to create an empty servemux.
	mux := http.NewServeMux()
	corsMux := middlewareCors(mux)

	mux.Handle("/", http.FileServer(http.Dir(".")))
	server := &http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}
	server.ListenAndServe()
	fmt.Println(server.Addr)
	fmt.Println("https://localhost:8080")
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
