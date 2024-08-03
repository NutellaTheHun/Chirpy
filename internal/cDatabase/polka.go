package cDatabase

import (
	"internal/api"
	"log"
	"net/http"
	"strings"
)

type PolkaRequest struct {
	Event string `json:"event"`
	Data  Data
}

type Data struct {
	UserId int `json:"user_id"`
}

func (db *DB) HandlePolkaPostWebHook(w http.ResponseWriter, r *http.Request) {

	header := r.Header.Get("Authorization")
	AuthString := strings.Split(header, " ")
	if len(AuthString) < 2 {
		w.WriteHeader(401)
		return
	}
	key := AuthString[1]
	if key != db.polkaApi {
		w.WriteHeader(401)
		return
	}

	var requestBody PolkaRequest
	err := api.RecieveJson(w, r, &requestBody)
	if err != nil {
		log.Print("PolkaPostWH, RecieveJson:", err.Error())
		w.WriteHeader(400)
		return
	}

	if requestBody.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	if requestBody.Event == "user.upgraded" {
		err = db.UpgradeUser(requestBody.Data.UserId)
		if err != nil {
			if err.Error() == "404" {
				w.WriteHeader(404)
				return
			}
		}
		w.WriteHeader(204)
	}
}
