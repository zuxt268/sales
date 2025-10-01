package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func SlogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			err := next(c)

			slog.Info("HTTP request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", res.Status,
				"latency", time.Since(start).String(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
			)

			return err
		}
	}
}
