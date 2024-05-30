package main

import (
	"encoding/json"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var CHIRPID int = 1
var USERID int = 1

type DB struct {
	Path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PublicUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func createDB(path string) (*DB, error) {
	db := DB{Path: path, mux: &sync.RWMutex{}}

	if _, err := os.Stat(path); err == nil {
		return &db, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	db.writeDB(DBStructure{Chirps: make(map[int]Chirp), Users: make(map[int]User)})
	return &db, nil
}

func deleteDB(path string) {
	os.Remove(path)
}

func (db *DB) addChirp(chirp Chirp) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	data.Chirps[CHIRPID] = chirp
	db.writeDB(data)
	CHIRPID++
}

func (db *DB) addUser(user User) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	data.Users[USERID] = user
	passByte, err := bcrypt.GenerateFromPassword([]byte(user.Password), 0)
	if err != nil {
		panic(err)
	}
	user.Password = string(passByte)
	db.writeDB(data)
	USERID++
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

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var dbStruct DBStructure
	json.Unmarshal(f, &dbStruct)

	return dbStruct, nil
}

func (db *DB) getChirps() []Chirp {
	db.mux.RLock()
	defer db.mux.RUnlock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	var chirpMap DBStructure
	json.Unmarshal(f, &chirpMap)
	var allChirps []Chirp
	for _, chirp := range chirpMap.Chirps {
		allChirps = append(allChirps, chirp)
	}
	return allChirps
}
