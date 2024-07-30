package cDatabase

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

/*
	os.ReadFile
	os.ErrNotExist
	os.WriteFile
	sort.slice
*/

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(fpath string) (*DB, error) {

	db := &DB{path: fpath, mux: &sync.RWMutex{}}
	db.ensureDB()
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	id := len(dbStruct.Chirps) + 1
	result := Chirp{Id: id, Body: body}
	dbStruct.Chirps[id] = result

	db.writeDB(dbStruct)
	return result, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	var result []Chirp

	dbStruct, err := db.loadDB()
	if err != nil {
		return result, err
	}

	for _, item := range dbStruct.Chirps {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Id < result[j].Id })
	return result, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	os.Remove(db.path)
	os.Create(db.path)
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStruct := DBStructure{
		Chirps: make(map[int]Chirp),
	}

	data, err := os.ReadFile(db.path)
	if err != nil {
		return dbStruct, err
	}

	var chirps []Chirp
	err = json.Unmarshal(data, &chirps)
	if err != nil {
		return dbStruct, err
	}

	for _, item := range chirps {
		dbStruct.Chirps[item.Id] = item
	}
	return dbStruct, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	chirps, err := db.GetChirps()
	if err != nil {
		return err
	}
	data, err := json.Marshal(chirps)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) HandleGetChirpsRequest(w http.ResponseWriter, r *http.Request) {

}
func (db *DB) HandlePostChirpsRequest(w http.ResponseWriter, r *http.Request) {

}
