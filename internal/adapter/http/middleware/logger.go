package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestLogger returns an Echo middleware that emits a structured log line
// per request including method, path, status, latency, and request-id.
func RequestLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()

			fields := []zap.Field{
				zap.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", res.Status),
				zap.Duration("latency", time.Since(start)),
				zap.String("remote_ip", c.RealIP()),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				log.Warn("request completed with error", fields...)
			} else {
				log.Info("request completed", fields...)
			}

			return err
		}
	}
}
