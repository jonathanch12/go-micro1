package handler

import (
	"core-service/internal/model"

	"github.com/gin-gonic/gin"
)

// respondError writes a JSON error body matching the OpenAPI Error schema.
func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, model.ErrorResponse{Error: message})
}
