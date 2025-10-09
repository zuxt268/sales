package main

import (
	"context"
	"errors"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/di"
	"github.com/zuxt268/sales/internal/infrastructure"
	middleware2 "github.com/zuxt268/sales/internal/interfaces/middleware"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zuxt268/sales/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Sales API
// @version 1.0
// @description ドメイン管理API
// @BasePath /api
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// slog初期化
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting sales API server")

	// DB接続
	db := infrastructure.NewDatabase()

	// 依存性注入
	handler := di.Initialize(db)

	// Swagger hostを環境変数から設定
	docs.SwaggerInfo.Host = config.Env.SwaggerHost

	e := echo.New()

	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!!")
	})

	api := e.Group("/api")
	api.Use(middleware2.JWTMiddleware())
	api.Use(middleware2.SlogMiddleware())

	api.GET("/domains", handler.GetDomains)
	api.PUT("/domains/:id", handler.UpdateDomain)
	api.DELETE("/domains/:id", handler.DeleteDomain)
	api.POST("/fetch", handler.FetchDomains)
	api.POST("/domains/analyze", handler.AnalyzeDomains)

	srv := &http.Server{
		Addr:    config.Env.Address,
		Handler: e,
	}

	go func() {
		slog.Info("HTTP server started", "address", config.Env.Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	slog.Info("Shutting down server...")

	conn, err := db.DB()
	if err == nil {
		_ = conn.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
