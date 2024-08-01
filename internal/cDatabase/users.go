package cDatabase

import (
	"crypto/rand"
	"encoding/hex"
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
	Id           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	Token string `json:"token"`
}

type CustomClaim struct {
	jwt.RegisteredClaims
}

func CreateRegisteredClaims(id, expireTime int) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Second)),
		Subject:   strconv.Itoa(id),
	}
}

func CreateJWTAuthToken(id, expireTime int) *jwt.Token {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		CreateRegisteredClaims(id, expireTime),
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
		log.Print("createUser, getUserByEmail:", err.Error())
		return User{}, err
	}

	pword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Print("createUser, generatePasswordErr:", err.Error())
		return User{}, err
	}
	c := 4
	b := make([]byte, c)
	r, err := rand.Read(b)
	if err != nil {
		log.Print("createUser, rand.Read error:", err.Error())
	}

	//Create user
	id := len(dbStruct.Users) + 1
	newUser := User{Id: id, Email: email, Password: string(pword)}
	dbStruct.Users[id] = newUser

	//Create Refresh Token
	rTokenString := hex.EncodeToString([]byte(strconv.Itoa(r)))
	expireAt := time.Now().Add(time.Duration(60*24) * time.Hour)
	dbStruct.RTokens[id] = RToken{Token: rTokenString, ExpireAt: expireAt}

	//save
	db.writeDB(dbStruct)

	return newUser, nil
}

func (db *DB) updateUser(id int, userRequest *UserRequest) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Print("updateUser, loadDB error:", err.Error())
		return User{}, err
	}
	modUser := dbStruct.Users[id]
	modUser.Email = userRequest.Email
	pWord, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), 10)
	if err != nil {
		log.Print("updateUser, generatePasswordErr:", err.Error())
		return User{}, err
	}
	modUser.Password = string(pWord)
	dbStruct.Users[id] = modUser
	db.writeDB(dbStruct)

	return modUser, nil
}

// {email, password} -> {email, id}
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

// {email, password, expires_in_seconds} -> {id, email, token}
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
	//log.Print("REQUEST EXPIRE TIME:", request.ExpireTime)
	t := CreateJWTAuthToken(user.Id, 2)
	s, err := t.SignedString([]byte(db.secret))
	if err != nil {
		log.Print("SIGN ERROR", err.Error())
	}

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Print("Post Login, loadDB: ", err.Error())
		w.WriteHeader(401)
		return
	}
	userResp := UserLoginResponse{Id: user.Id, Email: user.Email, Token: s, RefreshToken: dbStruct.RTokens[user.Id].Token}
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

// { H{"Authorization: ${jwtToken}"}, {email, password} } -> {email, id}
func (db *DB) HandlePutUsersRequest(w http.ResponseWriter, r *http.Request) {

	//Parse Request Payload
	var request UserRequest
	err := api.RecieveJson(w, r, &request)
	if err != nil {
		log.Print(err.Error())
		return
	}

	//Get token
	header := r.Header.Get("Authorization")
	AuthString := strings.Split(header, " ")
	tokenString := AuthString[1]

	//Parse token string
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(db.secret), nil
	})
	if err != nil {
		log.Print("Put Users Parse With Claims: ", err.Error())
		w.WriteHeader(401)
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		log.Println("token.Claims not valid")
		w.WriteHeader(401)
		return
	}

	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		log.Print("id string to int error", err.Error())
		w.WriteHeader(401)
		return
	}

	//update database
	user, err := db.updateUser(id, &request)
	if err != nil {
		log.Print("Put Users, updateUser error:", err.Error())
		w.WriteHeader(401)
	}

	//trim pword
	response := UserResponse{Id: user.Id, Email: user.Email}

	//send response
	err = api.SendJson(w, r, response, 200)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(401)
		return
	}
}

func (db *DB) HandlePostRefresh(w http.ResponseWriter, r *http.Request) {
	//No request body, Referesh Token String in header "Authorization: Bearer <token>"
	AuthVal := r.Header.Get("Authorization")
	split := strings.Split(AuthVal, " ")
	tokenString := split[1]

	//Lookup token in DB
	result, err := CreateAccessToken(tokenString)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	//401 of NE or expire

	//200 return 1hr access token "token" : ${tokenString}
}

func (db *DB) CreateAccessToken(rToken string) (string, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Print("CreateAccessToken loadDB:", err.Error())
		return "", err
	}
}

func (db *DB) HandlePostRevoke(w http.ResponseWriter, r *http.Request) {
	//No request body, refresh token string in header  "Authorization: Bearer <token>"

	//Remove token from DB

	//return 204
}
