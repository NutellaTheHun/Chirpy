module github.com/NutellaTheHun/chirpy

go 1.22.5

require (
    internal/api v0.0.0
    internal/cDatabase v0.0.0
)
replace (
	internal/api v0.0.0 => ./internal/api
	internal/cDatabase v0.0.0 => ./internal/cDatabase
)
