package cDatabase

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type response struct {
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := os.ReadFile(db.path)
	if err != nil {
		log.Printf("loadDB readfile error")
		return DBStructure{}, err
	}

	if len(dat) == 0 {
		return DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}, nil
	}

	var result DBStructure
	err = json.Unmarshal(dat, &result)
	if err != nil {
		log.Printf("loadDB UnMarshall error")
		return DBStructure{}, err
	}

	return result, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, []byte(dat), 0666)
	if err != nil {
		return err
	}

	return nil
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(fpath string) (*DB, error) {
	db := &DB{path: fpath, mux: &sync.RWMutex{}}
	db.ensureDB()
	return db, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	os.Remove(db.path)
	os.Create(db.path)
	return nil
}
