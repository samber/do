package main

import (
	"github.com/labstack/echo/v4"
	echohttp "github.com/samber/do/v2/http/echo"
)

func main() {
	injector := startProgram()

	router := echo.New()
	echohttp.Use(router.Group("/debug/do"), "/debug/do", injector)

	router.Start(":8080")
}
