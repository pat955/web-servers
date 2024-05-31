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
	Chirps map[int]Chirp   `json:"chirps"`
	Users  map[string]User `json:"users"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID               int    `json:"id"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type PublicUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (u *User) userToPublic() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email}
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
	db.writeDB(DBStructure{Chirps: make(map[int]Chirp), Users: make(map[string]User)})
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
	passByte, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		panic(err)
	}
	user.Password = string(passByte)

	data.Users[user.Email] = user
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

func (db *DB) getUsers() []User {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var data DBStructure
	json.Unmarshal(f, &data)
	var allUsers []User
	for _, user := range data.Users {
		allUsers = append(allUsers, user)
	}
	return allUsers
}
func (db *DB) getChirpMap() map[int]Chirp {
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

func (db *DB) getUsersMap() map[string]User {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var data DBStructure
	json.Unmarshal(f, &data)
	return data.Users
}
