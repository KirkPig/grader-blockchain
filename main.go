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
	router.POST("/api/v1/submit", handler.SentCodeHandler)
	router.POST("/api/v1/check", handler.CheckCodeHandler)
	router.POST("/api/v1/lost", handler.ChangeKeyHandler)
	router.POST("/api/v1/close", handler.CloseSystemHandler)

	router.Run("localhost:1323")
}
