package handler

import (
	"errors"
	"net/http"

	"ewallet/internal/model"
	"ewallet/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.users.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			respondError(c, http.StatusConflict, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to register user")
		return
	}

	c.JSON(http.StatusCreated, model.UserRegisterResponse{
		UserID:  user.ID.String(),
		Name:    user.Name,
		Email:   user.Email,
		Message: "User registered successfully",
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.users.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondError(c, http.StatusUnauthorized, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to authenticate")
		return
	}

	c.JSON(http.StatusOK, model.UserLoginResponse{Token: token})
}
