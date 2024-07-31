module github.com/NutellaTheHun/chirpy

go 1.22.5

require (
	internal/api v0.0.0
	internal/cDatabase v0.0.0
)

require golang.org/x/crypto v0.25.0 // indirect

replace (
	internal/api v0.0.0 => ./internal/api
	internal/cDatabase v0.0.0 => ./internal/cDatabase
)
