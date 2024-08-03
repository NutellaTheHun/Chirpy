package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Accept POST /api/validate_chirp
type JsonBody struct {
	Body string `json:"body"`
}

// return and code 400
type JsonErr struct {
	ErrorMsg string `json:"error"`
}

// return and code 200
type JsonValid struct {
	ValidMsg string `json:"cleaned_body"`
}

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
	request := &JsonBody{}

	err := RecieveJson(w, r, request)
	if err != nil {
		SendJson(w, r, JsonErr{ErrorMsg: "Something went wrong"}, 400)
		return
	}
	if len(request.Body) > 140 {
		SendJson(w, r, JsonErr{ErrorMsg: "Chirp is too long"}, 400)
		return
	}
	SendJson(w, r, JsonValid{ValidMsg: CleanMsg(request.Body)}, 200)
}

func CleanMsg(inputMsg string) string {
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

func RecieveJson(w http.ResponseWriter, r *http.Request, s interface{}) error {
	decoder := json.NewDecoder(r.Body)
	result := s
	err := decoder.Decode(&result)
	if err != nil {
		return errors.New("recieve json decode error")
	}
	return nil
}

func SendJson(w http.ResponseWriter, r *http.Request, s interface{}, statusCode int) error {
	response, err := json.Marshal(s)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(response)
	if err != nil {
		return err
	}
	return nil
}
