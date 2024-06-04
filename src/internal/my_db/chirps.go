package my_db

import (
	"encoding/json"
	"os"
	"regexp"
)

type Chirp struct {
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
	Body     string `json:"body"`
}

func Censor(s string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	return re.ReplaceAllString(s, "****")
}

func (db *DB) AddChirp(chirp Chirp) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	data.Chirps[chirp.ID] = chirp
	chirp.Body = Censor(chirp.Body)
	db.writeDB(data)
}
func (db *DB) GetChirps() []Chirp {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var data DBStructure
	json.Unmarshal(f, &data)
	var allChirps []Chirp
	for _, chirp := range data.Chirps {
		allChirps = append(allChirps, chirp)
	}
	return allChirps
}

func (db *DB) GetChirpMap() map[int]Chirp {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var data DBStructure
	json.Unmarshal(f, &data)
	return data.Chirps
}

func (db *DB) GetChirp(id int) (Chirp, bool) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	chirp, found := data.Chirps[id]
	if !found {
		return Chirp{}, false
	}
	return chirp, true
}
