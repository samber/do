package main

import (
	"github.com/gofiber/fiber/v2"
	fiberhttp "github.com/samber/do/http/fiber/v2"
)

func main() {
	injector := startProgram()

	router := fiber.New()
	fiberhttp.Use(router.Group("/debug/do"), "/debug/do", injector)

	router.Listen(":8080")
}
