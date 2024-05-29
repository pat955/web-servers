package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var NEWID int = 0

type DB struct {
	Path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chrips map[int]Chirp `json:"chirps"`
}
type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func createDB(path string) (*DB, error) {
	fmt.Println("creating db")
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	db := DB{Path: path, mux: &sync.RWMutex{}}
	db.writeDB(DBStructure{})
	return &db, nil
}

func deleteDB(path string) {
	fmt.Println("deleting db")
	err := os.Remove(path)
	if err != nil {
		panic(err)
	}
}

func (db *DB) createChirp(body string) (Chirp, error) {
	if 140 >= len(body) && len(body) >= 1 {
		newChirp := Chirp{Id: NEWID, Body: body}
		NEWID++
		chirpMap, err := db.loadDB()
		if err != nil {
			panic(err)
		}
		fmt.Println(chirpMap.Chrips)
		chirpMap.Chrips[NEWID] = newChirp
		db.writeDB(chirpMap)
		return newChirp, nil
	} else {
		return Chirp{}, errors.New("invalid chirp")
	}
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
	fmt.Println(dbStruct)

	return dbStruct, nil
}
func (db *DB) getChirps() map[int]Chirp {
	db.mux.RLock()
	defer db.mux.RUnlock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	var chirpMap DBStructure
	json.Unmarshal(f, &chirpMap)
	return chirpMap.Chrips
}
