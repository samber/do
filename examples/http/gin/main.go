package main

import (
	"github.com/gin-gonic/gin"
	ginhttp "github.com/samber/do/http/gin/v2"
)

func main() {
	injector := startProgram()

	router := gin.New()
	ginhttp.Use(router.Group("/debug/do"), injector)

	router.Run(":8088")
}
