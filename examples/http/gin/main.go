package main

import (
	"github.com/gin-gonic/gin"
	ginhttp "github.com/samber/do/v2/http/gin"
)

func main() {
	injector := startProgram()

	router := gin.New()
	ginhttp.Use(router.Group("/debug/do"), injector)

	router.Run(":8080")
}
