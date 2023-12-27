package main

import (
	"net/http"

	"github.com/samber/do/http/std"
)

func main() {
	injector := startProgram()

	mux := http.NewServeMux()
	mux.Handle("/debug/do/", std.Use("/debug", injector))

	http.ListenAndServe(":8080", mux)
}
