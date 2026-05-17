package router

import (
	"github.com/eghbalii/go-company-service/internal/adapter/http/handler"
	mw "github.com/eghbalii/go-company-service/internal/adapter/http/middleware"
	"github.com/eghbalii/go-company-service/pkg/jwt"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Register wires all routes onto the Echo instance.
func Register(
	e *echo.Echo,
	jwtMgr *jwt.Manager,
	authH *handler.AuthHandler,
	companyH *handler.CompanyHandler,
) {
	// Swagger UI — served at /swagger/index.html
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")

	// Public routes
	api.POST("/auth/login", authH.Login)

	// Company read — public
	api.GET("/companies/:id", companyH.GetByID)

	// Company mutations — protected
	protected := api.Group("", mw.JWT(jwtMgr))
	protected.POST("/companies", companyH.Create)
	protected.PATCH("/companies/:id", companyH.Update)
	protected.DELETE("/companies/:id", companyH.Delete)
}
