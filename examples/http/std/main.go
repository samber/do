package main

import (
	"net/http"

	// "github.com/samber/do/http/std/v2"
	"github.com/samber/do/v2/http/std"
)

func main() {
	injector := startProgram()

	mux := http.NewServeMux()
	mux.Handle("/debug/do/", std.Use("/debug/do", injector))

	http.ListenAndServe(":8080", mux)
}
