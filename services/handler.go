package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) GetTransactionHandler(c *gin.Context) {

	pubKey := c.Param("pub_key")

	transaction_list, err := h.service.GetAllTransaction(pubKey)

	if err != nil {
		c.IndentedJSON(http.StatusOK, &gin.H{
			"log": err,
		})
		return
	}

	c.IndentedJSON(http.StatusOK, transaction_list)

}

func (h *Handler) AuthorizationHandler(c *gin.Context) {

	var req AuthorizationRequest

	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	response, err := h.service.Authorization(&req)
	if err != nil {
		log.Println(err)
	}
	c.JSON(http.StatusOK, response)

}
