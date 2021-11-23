package main

import (
	"github.com/KirkPig/grader-blockchain/services"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	handler := services.NewHandler(services.NewService())

	router.POST("/api/v1/authorization/new", handler.AuthorizationHandler)

	router.Run("localhost:1323")
}
