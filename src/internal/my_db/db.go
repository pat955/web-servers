package my_db

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

type DB struct {
	Path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RefreshTokens map[string]TokenInfo `json:"refresh_tokens"`
}

func CreateDB(path string) *DB {
	db := DB{Path: path, mux: &sync.RWMutex{}}

	if _, err := os.Stat(path); err == nil {
		return &db
	}
	f, _ := os.Create(path)

	defer f.Close()
	db.writeDB(DBStructure{Chirps: make(map[int]Chirp), Users: make(map[int]User), RefreshTokens: make(map[string]TokenInfo)})
	return &db
}

func DeleteDB(path string) {
	os.Remove(path)
}

func (db *DB) writeDB(dbstruct DBStructure) {
	json, err := json.Marshal(dbstruct)
	if err != nil {
		panic(err)
	}
	db.mux.Lock()
	os.WriteFile(db.Path, json, os.ModePerm)
	db.mux.Unlock()
}

func (db *DB) loadDB() DBStructure {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var dbStruct DBStructure
	json.Unmarshal(f, &dbStruct)
	return dbStruct
}

// decodes json into your provided struct. Using this to avoid making a massive all encompassing struct
func DecodeForm(req *http.Request, dst interface{}) {
	if err := json.NewDecoder(req.Body).Decode(dst); err != nil {
		panic(err)
	}
}
