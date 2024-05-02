package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerCount(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf("<html>\n\n	<body>\n		<h1>Welcome, Chirpy Admin</h1>\n		<p>Chirpy has been visited %d times!</p>\n	</body>\n\n	</html>", cfg.fileserverHits)
	w.Write([]byte(fmt.Sprint(html)))
}

func (cfg *apiConfig) handlerResetCount(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits)))
}
