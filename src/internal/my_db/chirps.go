package my_db

import (
	"errors"
	"regexp"
)

type Chirp struct {
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
	Body     string `json:"body"`
}

func (db *DB) AddChirp(chirp Chirp) {
	data := db.loadDB()
	data.Chirps[chirp.ID] = chirp
	chirp.Body = censor(chirp.Body)
	db.writeDB(data)
}

func (db *DB) GetChirps() []Chirp {
	data := db.loadDB()

	var allChirps []Chirp
	for _, chirp := range data.Chirps {
		allChirps = append(allChirps, chirp)
	}
	return allChirps
}

func (db *DB) GetChirp(id int) (Chirp, bool) {
	data := db.loadDB()
	chirp, found := data.Chirps[id]
	if !found {
		return Chirp{}, false
	}
	return chirp, true
}

func (db *DB) DeleteChirp(chirpID, authorID int) error {
	data := db.loadDB()
	if data.Chirps[chirpID].AuthorID == authorID {
		delete(data.Chirps, data.Chirps[chirpID].ID)
		return nil
	}
	return errors.New("cannot delete chirp not writen by you")
}

func censor(s string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	return re.ReplaceAllString(s, "****")
}
