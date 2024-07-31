module internal/cDatabase

go 1.22.5

require internal/api v0.0.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	golang.org/x/crypto v0.25.0
)

replace internal/api v0.0.0 => ../api
