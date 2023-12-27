package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chihttp "github.com/samber/do/v2/http/chi"
)

func main() {
	injector := startProgram()

	router := chi.NewRouter()
	chihttp.Use(router, "/debug/do", injector)

	http.ListenAndServe(":8080", router)
}
