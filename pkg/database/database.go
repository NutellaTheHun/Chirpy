package database

type chirpRequest struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type dataBase struct {
	count  int
	chirps []chirpResponse
}
