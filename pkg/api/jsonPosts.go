package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Accept POST /api/validate_chirp
type jsonBody struct {
	Body string `json:"body"`
}

// return and code 400
type jsonErr struct {
	ErrorMsg string `json:"error"`
}

// return and code 200
type jsonValid struct {
	ValidMsg string `json:"cleaned_body"`
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	request := &jsonBody{}

	err := recieveJson(w, r, request)
	if err != nil {
		sendJson(w, r, jsonErr{ErrorMsg: "Something went wrong"}, 400)
		return
	}
	if len(request.Body) > 140 {
		sendJson(w, r, jsonErr{ErrorMsg: "Chirp is too long"}, 400)
		return
	}
	sendJson(w, r, jsonValid{ValidMsg: cleanMsg(request.Body)}, 200)
}

func cleanMsg(inputMsg string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	tokens := strings.Split(inputMsg, " ")
	for i, token := range tokens {
		for _, badWord := range badWords {
			if strings.ToLower(token) == badWord {
				tokens[i] = "****"
				break
			}
		}
	}
	return strings.Join(tokens[:], " ")
}

func recieveJson(w http.ResponseWriter, r *http.Request, s interface{}) error {
	decoder := json.NewDecoder(r.Body)
	result := s
	err := decoder.Decode(&result)
	if err != nil {
		return errors.New("Something went wrong")
	}
	return nil
}

func sendJson(w http.ResponseWriter, r *http.Request, s interface{}, statusCode int) error {
	response, err := json.Marshal(s)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
	return nil
}
