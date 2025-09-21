package main

import (
	"github.com/labstack/echo/v4"
	doecho "github.com/samber/do/http/echo/v2"
)

func main() {
	injector := startProgram()

	router := echo.New()
	doecho.Use(router.Group("/debug/do"), "/debug/do", injector)

	router.Start(":8080")
}
