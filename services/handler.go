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

	c.JSON(http.StatusOK, h.service.Authorization(&req))

}
