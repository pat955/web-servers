package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var CHIRPID int = 1
var USERID int = 1

type DB struct {
	Path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chrips map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
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
	db.writeDB(DBStructure{Chrips: make(map[int]Chirp), Users: make(map[int]User)})
	return &db, nil
}

func deleteDB(path string) {
	os.Remove(path)
}

func (db *DB) createChirp(body string) (Chirp, error) {
	newChirp := Chirp{ID: CHIRPID, Body: body}

	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	data.Chrips[CHIRPID] = newChirp
	db.writeDB(data)
	CHIRPID++
	return newChirp, nil

}

func (db *DB) createUser(email string) (User, error) {
	fmt.Println(email)

	newUser := User{ID: USERID, Email: email}
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	data.Users[USERID] = newUser
	db.writeDB(data)
	USERID++
	return newUser, nil
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
	for _, chirp := range chirpMap.Chrips {
		allChirps = append(allChirps, chirp)
	}
	return allChirps
}
