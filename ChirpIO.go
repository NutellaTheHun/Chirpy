package main

import (
	"net/http"
)

type chirpRequest struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type chirpCounter struct {
	count int
}

func GetChirp(w http.ResponseWriter, r *http.Request) {

}

func PostChirp(w http.ResponseWriter, r *http.Request) {
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

	sendJson(w, r, chirpResponse{Id: 0, Body: cleanMsg(request.Body)}, 201)
}
