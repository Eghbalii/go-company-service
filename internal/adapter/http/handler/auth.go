package handler

import (
	"net/http"

	"github.com/eghbalii/go-company-service/internal/adapter/http/dto"
	"github.com/eghbalii/go-company-service/internal/usecase/auth"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	uc auth.UseCase
}

// NewAuthHandler creates an AuthHandler backed by the given use case.
func NewAuthHandler(uc auth.UseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticates a user and returns a JWT token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest	true	"Credentials"
//	@Success		200		{object}	dto.SuccessResponse{data=dto.LoginResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	token, err := h.uc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{Token: token}))
}
