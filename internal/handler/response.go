package handler

import (
	"ewallet/internal/model"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, model.ErrorResponse{Error: message})
}
