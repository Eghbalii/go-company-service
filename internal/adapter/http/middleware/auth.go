package middleware

import (
	"net/http"
	"strings"

	"github.com/eghbalii/go-company-service/pkg/jwt"
	"github.com/labstack/echo/v4"
)

const userIDKey = "user_id"
const userEmailKey = "user_email"

// JWT returns an Echo middleware that validates Bearer tokens.
// On success it stores user_id and user_email in the Echo context.
func JWT(mgr *jwt.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if header == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "malformed authorization header")
			}

			claims, err := mgr.Verify(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set(userIDKey, claims.UserID)
			c.Set(userEmailKey, claims.Email)

			return next(c)
		}
	}
}
