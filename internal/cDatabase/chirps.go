package cDatabase

import (
	"errors"
	"internal/api"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
	chirps, err := db.GetChirps()
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	err = api.SendJson(w, r, chirps, 200)
	if err != nil {
		log.Print(err.Error())
	}
}

func (db *DB) HandlePostChirpsRequest(w http.ResponseWriter, r *http.Request) {
	var respBody response
	err := api.RecieveJson(w, r, &respBody)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(500)
	}

	//VALIDATE USER,
	//Get token
	header := r.Header.Get("Authorization")
	AuthString := strings.Split(header, " ")
	tokenString := AuthString[1]

	claims, err := db.DecodeJWTToken(tokenString)
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
		log.Print(err.Error())
		w.WriteHeader(500)
		return
	}
}
