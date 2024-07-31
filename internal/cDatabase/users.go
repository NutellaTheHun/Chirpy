package cDatabase

import (
	"encoding/json"
	"errors"
	"internal/api"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type UserLoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	ExpireTime int    `json:"expires_in_seconds"`
}

type UserLoginResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func CreateJWTAuthToken(id, expireTime int) *jwt.Token {

	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Second)),
			Subject:   strconv.Itoa(id),
		},
	)
	return t
}

func (db *DB) createUser(email, password string) (User, error) {

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Print("createUser, loadDB")
		return User{}, err
	}

	_, err = db.getUserByEmail(email)
	if err == nil {
		log.Print("createUser, getUserByEmail")
		return User{}, err
	}

	pword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		log.Print(err.Error())
	}

	id := len(dbStruct.Users) + 1
	newUser := User{Id: id, Email: email, Password: string(pword)}
	dbStruct.Users[id] = newUser
	db.writeDB(dbStruct)

	return newUser, nil
}

func (db *DB) HandlePostUsers(w http.ResponseWriter, r *http.Request) {

	var request UserRequest
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&request)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(500)
		return
	}

	user, err := db.createUser(request.Email, request.Password)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(500)
		return
	}

	//trim pword
	response := UserResponse{Id: user.Id, Email: user.Email}

	err = api.SendJson(w, r, response, 201)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(500)
		return
	}
}

func (db *DB) HandleLoginRequest(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var request UserLoginRequest

	err := decoder.Decode(&request)
	if err != nil {
		log.Print("handleLoginRequest ", err.Error())
		w.WriteHeader(500)
		return
	}

	user, err := db.getUserByEmail(request.Email)
	if err != nil {
		log.Print(err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		w.WriteHeader(401)
		return
	}

	t := CreateJWTAuthToken(user.Id, request.ExpireTime)
	s, err := t.SignedString([]byte(db.secret))
	if err != nil {
		log.Print("SIGN ERROR", err.Error())
	}

	userResp := UserLoginResponse{Id: user.Id, Email: user.Email, Token: s}
	err = api.SendJson(w, r, userResp, 200)
	if err != nil {
		w.WriteHeader(500)
		return
	}
}

func (db *DB) getUserByEmail(email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, errors.New("failed to load db")
	}

	for _, item := range dbStruct.Users {
		if item.Email == email {
			return item, nil
		}
	}
	return User{}, errors.New("User not found")
}

func (db *DB) HandlePutUsersRequest(w http.ResponseWriter, r *http.Request) {
	var response UserRequest
	err := api.RecieveJson(w, r, &response)
	if err != nil {
		log.Print(err.Error())
		return
	}
	header := r.Header.Get("Authorization")
	AuthString := strings.Split(header, " ")
	t := AuthString[1]
	//log.Print("PUT REQUEST HEADER TOKEN: ", t)
	valid, err := jwt.ParseWithClaims(t, jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(db.secret), nil
	})
	if err != nil {
		log.Print("Put Users Parse With Claims: ", err.Error())
		return
	}
	log.Print(valid.Claims.GetIssuer())
}
