package cDatabase

import (
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (db *DB) DecodeJWTToken(r *http.Request) (*jwt.RegisteredClaims, error) {

	header := r.Header.Get("Authorization")
	AuthString := strings.Split(header, " ")
	tokenString := AuthString[1]

	//Parse token string
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(db.secret), nil
	})
	if err != nil {
		log.Print("Put Users Parse With Claims: ", err.Error())
		return &jwt.RegisteredClaims{}, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		log.Println("token.Claims not valid")
		return &jwt.RegisteredClaims{}, err
	}
	return claims, nil
}
