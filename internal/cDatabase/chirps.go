package cDatabase

import (
	"errors"
	"internal/api"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	id := len(dbStruct.Chirps) + 1
	result := Chirp{Id: id, Body: body, AuthorId: userId}
	dbStruct.Chirps[id] = result

	db.writeDB(dbStruct)

	return result, nil
}

// GetChirps returns all chirps in the database
// idQuery optional
func (db *DB) GetChirps(idQuery string) ([]Chirp, error) {
	var result []Chirp

	id := 0
	if idQuery != "" {
		result, err := strconv.Atoi(idQuery)
		if err != nil {
			return []Chirp{}, errors.New("GetChirps id conv -> int error")
		}
		id = result
	}

	dbStruct, err := db.loadDB()
	if err != nil {
		return result, err
	}

	if id == 0 {
		for _, item := range dbStruct.Chirps {
			result = append(result, item)
		}
	} else {
		for _, item := range dbStruct.Chirps {
			if item.AuthorId == id {
				result = append(result, item)
			}
		}
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

func (db *DB) HandleGetChirpRequest(w http.ResponseWriter, r *http.Request) {

	pathVal := r.PathValue("chirpId")
	id, err := strconv.Atoi(pathVal)
	if err != nil {
		log.Print("getChirpById, strconv err ", err.Error())
	}

	chirp, err := db.GetChirp(id)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	err = api.SendJson(w, r, chirp, 200)
	if err != nil {
		log.Print("getChirpById, SendJSON ", err.Error())
	}
}

func (db *DB) HandleGetChirpsRequest(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")
	chirps, err := db.GetChirps(id)
	if err != nil {
		log.Fatal(err)
	}
	if sortQuery == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id > chirps[j].Id })
	} else {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	}

	err = api.SendJson(w, r, chirps, 200)
	if err != nil {
		log.Print("GetChirps SendJson", err.Error())
	}
}

func (db *DB) HandlePostChirpsRequest(w http.ResponseWriter, r *http.Request) {

	//VALIDATE USER,
	claims, err := db.DecodeJWTToken(r)
	if err != nil {
		log.Print("PostChirps, DecodeToken:", err.Error())
		w.WriteHeader(400)
		return
	}

	if claims.ExpiresAt.Before(time.Now()) {
		log.Print("POST Chirps, JWT Token expired")
		w.WriteHeader(400)
		return
	}

	var respBody response
	err = api.RecieveJson(w, r, &respBody)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(500)
	}

	authorId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		log.Print("POST Chirps, strconv uId -> int", err.Error())
		w.WriteHeader(400)
		return
	}

	chirp, err := db.CreateChirp(respBody.Body, authorId) //Pass user id
	if err != nil {
		log.Fatal(err)
	}

	err = api.SendJson(w, r, chirp, 201)
	if err != nil {
		log.Print("postChirp sendJson:", err.Error())
		w.WriteHeader(500)
		return
	}
}

func (db *DB) HandleDeleteChirpsRequest(w http.ResponseWriter, r *http.Request) {

	claims, err := db.DecodeJWTToken(r)
	if err != nil {
		log.Print("DeleteChirpRequest, DecodeJTW:", err.Error())
		w.WriteHeader(400)
		return
	}

	claimId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		log.Print("DELETE chirp request, str conv subject -> int", err.Error())
		w.WriteHeader(400)
		return
	}

	pathVal := r.PathValue("chirpId")
	chirId, err := strconv.Atoi(pathVal)
	if err != nil {
		log.Print("getChirpById, strconv err ", err.Error())
		w.WriteHeader(400)
		return
	}

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Print("del chirp req, load db", err.Error())
		w.WriteHeader(400)
		return
	}
	chirp, ok := dbStruct.Chirps[chirId]
	if !ok {
		log.Print("Chirp id not found")
		w.WriteHeader(400)
		return
	}

	if chirp.AuthorId != claimId {
		log.Print("Chirp id not found")
		w.WriteHeader(403)
		return
	}
	db.DeleteChirp(chirId)
	w.WriteHeader(204)
}

func (db *DB) DeleteChirp(chirpId int) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Print("DeleteChirp(), load db", err.Error())
		return
	}
	delete(dbStruct.Chirps, chirpId)
	db.writeDB(dbStruct)
}
