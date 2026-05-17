package server

import (
	"net/http"

	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// customValidator adapts go-playground/validator to Echo's Validator interface.
type customValidator struct {
	v *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}

// New creates and configures an Echo instance.
func New(log *zap.Logger, debug bool) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = debug

	e.Validator = &customValidator{v: validator.New()}

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	e.HTTPErrorHandler = customErrorHandler(log)

	return e
}

// customErrorHandler maps apperr sentinels and Echo errors to consistent JSON responses.
func customErrorHandler(log *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		msg := "internal server error"

		switch {
		case apperr.Is(err, apperr.ErrNotFound):
			code, msg = http.StatusNotFound, err.Error()
		case apperr.Is(err, apperr.ErrConflict):
			code, msg = http.StatusConflict, err.Error()
		case apperr.Is(err, apperr.ErrValidation):
			code, msg = http.StatusUnprocessableEntity, err.Error()
		case apperr.Is(err, apperr.ErrUnauthorized), apperr.Is(err, apperr.ErrInvalidCredential):
			code, msg = http.StatusUnauthorized, err.Error()
		case apperr.Is(err, apperr.ErrForbidden):
			code, msg = http.StatusForbidden, err.Error()
		default:
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
				if m, ok := he.Message.(string); ok {
					msg = m
				}
			} else {
				log.Error("unhandled error", zap.Error(err))
			}
		}

		_ = c.JSON(code, map[string]string{"error": msg})
	}
}
