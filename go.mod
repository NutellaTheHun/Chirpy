module github.com/NutellaTheHun/chirpy

go 1.22.5

require (
	internal/api v0.0.0
	internal/cDatabase v0.0.0
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/crypto v0.25.0 // indirect
)

replace (
	internal/api v0.0.0 => ./internal/api
	internal/cDatabase v0.0.0 => ./internal/cDatabase
)
