package cDatabase

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
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

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(fpath string) (*DB, error) {

	db := &DB{path: fpath, mux: &sync.RWMutex{}}
	db.ensureDB()
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
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
		log.Printf("loadDB empty return")
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
	log.Printf("loadDB nonempty return")
	return result, nil
	/*
	   	dbStruct := DBStructure{
	   		Chirps: make(map[int]Chirp),
	   		Users:  make(map[int]User),
	   	}

	   data, err := os.ReadFile(db.path)

	   	if err != nil {
	   		log.Printf("loadDB readfile")
	   		return dbStruct, err
	   	}

	   	if len(data) == 0 {
	   		log.Printf("loadDB empty return")
	   		return dbStruct, nil
	   	}

	   chirps := []Chirp{}
	   err = json.Unmarshal(data, &chirps)

	   	if err != nil {
	   		log.Printf("loadDB chirps unmarshal")
	   		return dbStruct, err
	   	}

	   	for _, item := range chirps {
	   		dbStruct.Chirps[item.Id] = item
	   	}

	   users := []User{}
	   err = json.Unmarshal(data, &users)

	   	if err != nil {
	   		log.Printf("loadDB Users unmarshal")
	   		return dbStruct, err
	   	}

	   	for _, item := range users {
	   		dbStruct.Users[item.Id] = item
	   	}

	   return dbStruct, nil
	*/
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
	/*
		chirps := []Chirp{}
		for _, item := range dbStructure.Chirps {
			chirps = append(chirps, item)
		}
		data, err := json.Marshal(chirps)
		if err != nil {
			return err
		}

		users := []User{}
		for _, item := range dbStructure.Users {
			users = append(users, item)
		}
		uData, err := json.Marshal(users)
		if err != nil {
			return err
		}
		db.mux.Lock()
		defer db.mux.Unlock()

		err = os.WriteFile(db.path, data, 0666)
		if err != nil {
			return err
		}
		err = os.WriteFile(db.path, uData, 0666)
		if err != nil {
			return err
		}*/
	return nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	var result []Chirp

	dbStruct, err := db.loadDB()
	if err != nil {
		return result, err
	}

	for _, item := range dbStruct.Chirps {
		result = append(result, item)
	}

	return result, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirp(id int) (Chirp, error) {

	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := dbStruct.Chirps[id]
	if ok {
		return chirp, nil
	}

	return Chirp{}, errors.New("id ${id} not found")
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	os.Remove(db.path)
	os.Create(db.path)
	return nil
}

func (db *DB) HandleGetChirpRequest(w http.ResponseWriter, r *http.Request) {
	//decoder := json.NewDecoder(r.Body)
	//var respBody response
	pathVal := r.PathValue("chirpId")
	id, err := strconv.Atoi(pathVal)
	if err != nil {
		log.Printf("getChirpById, strconv err ", err.Error())
	}
	chirp, err := db.GetChirp(id)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	dat, err := json.Marshal(chirp)
	if err != nil {
		log.Fatal("getChirpById, marshal ", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (db *DB) HandleGetChirpsRequest(w http.ResponseWriter, r *http.Request) {
	chirps, err := db.GetChirps()
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	dat, err := json.Marshal(chirps)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}
func (db *DB) HandlePostChirpsRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var respBody response
	err := decoder.Decode(&respBody)
	if err != nil {
		log.Printf("decoded body: ", respBody.Body)
		log.Printf("decoder.Decode: ", err.Error())
		w.WriteHeader(500)
		return
	}
	log.Printf("TRUE decoded body: ", respBody.Body)

	chirp, err := db.CreateChirp(respBody.Body)
	if err != nil {
		log.Fatal(err)
	}

	dat, err := json.Marshal(chirp)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
	return
}
