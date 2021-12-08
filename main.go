package main

import (
	"github.com/KirkPig/grader-blockchain/services"
	"github.com/gin-gonic/gin"
)

/*
TODO: Setup MySQL repository access
*/

func main() {
	router := gin.Default()
	handler := services.NewHandler(services.NewService())

	router.POST("/api/v1/authorization/new", handler.AuthorizationHandler)
	router.GET("/api/v1/transaction/:pub_key", handler.GetTransactionHandler)

	router.Run("localhost:1323")
}
