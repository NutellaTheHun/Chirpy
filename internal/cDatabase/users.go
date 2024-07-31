package cDatabase

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (db *DB) createUser(email, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Printf("createUser, loadDB")
		return User{}, err
	}

	_, err = db.getUserByEmail(email)
	if err == nil {
		log.Printf("createUser, getUserByEmail")
		return User{}, err
	}

	id := len(dbStruct.Users) + 1
	pword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	newUser := User{Id: id, Email: email, Password: string(pword)}
	dbStruct.Users[id] = newUser
	db.writeDB(dbStruct)
	return newUser, nil
}

func (db *DB) HandlePostUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request UserRequest
	err := decoder.Decode(&request)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(500)
		return
	}
	user, err := db.createUser(request.Email, request.Password)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(500)
		return
	}

	//trim pword
	response := UserResponse{Id: user.Id, Email: user.Email}

	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func (db *DB) HandleLoginRequest(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var request UserRequest

	err := decoder.Decode(&request)
	if err != nil {
		log.Printf("handleLoginRequest ", err.Error())
		w.WriteHeader(500)
		return
	}
	log.Printf("handleLoginRequest decode")
	user, err := db.getUserByEmail(request.Email)
	if err != nil {

	}
	log.Printf("handleLoginRequest getSUerEmail")
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		w.WriteHeader(401)
		return
	}
	log.Printf("handleLoginRequest compareHash")
	userResp := UserResponse{Id: user.Id, Email: user.Email}
	dat, err := json.Marshal(userResp)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	log.Printf("handleLoginRequest marshall")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (db *DB) getUserByEmail(email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Printf("getUserByEmail, loaddb failed")
		return User{}, errors.New("failed to load db")
	}

	for _, item := range dbStruct.Users {
		if item.Email == email {
			log.Printf("getUserByEmail, email found")
			return item, nil
		}
	}
	log.Printf("getUserByEmail, email not found")
	return User{}, errors.New("User not found")
}
