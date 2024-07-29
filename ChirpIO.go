package main

import (
	"net/http"
)

func (db *dataBase) GetChirp(w http.ResponseWriter, r *http.Request) {

}

func (db *dataBase) PostChirp(w http.ResponseWriter, r *http.Request) {
	request := &chirpResponse{}
	err := recieveJson(w, r, request)
	if err != nil {
		sendJson(w, r, jsonErr{ErrorMsg: "Something went wrong"}, 400)
		return
	}
	if len(request.Body) > 140 {
		sendJson(w, r, jsonErr{ErrorMsg: "Chirp is too long"}, 400)
		return
	}

	sendJson(w, r, database.chirpResponse{Id: 0, Body: cleanMsg(request.Body)}, 201)
}
