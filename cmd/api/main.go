// @title			Company Service API
// @version		1.0
// @description	Production-ready REST microservice for managing companies.
//
// @contact.name	Eghbali
//
// @license.name	MIT
//
// @host		localhost:8080
// @BasePath	/api/v1
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and the JWT token.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/eghbalii/go-company-service/docs"
	eventadapter "github.com/eghbalii/go-company-service/internal/adapter/event"
	"github.com/eghbalii/go-company-service/internal/adapter/http/handler"
	httpmw "github.com/eghbalii/go-company-service/internal/adapter/http/middleware"
	"github.com/eghbalii/go-company-service/internal/adapter/http/router"
	"github.com/eghbalii/go-company-service/internal/adapter/repository"
	"github.com/eghbalii/go-company-service/internal/infrastructure/database"
	infrakafka "github.com/eghbalii/go-company-service/internal/infrastructure/kafka"
	"github.com/eghbalii/go-company-service/internal/infrastructure/server"
	"github.com/eghbalii/go-company-service/internal/usecase/auth"
	"github.com/eghbalii/go-company-service/internal/usecase/company"
	"github.com/eghbalii/go-company-service/pkg/config"
	"github.com/eghbalii/go-company-service/pkg/jwt"
	"github.com/eghbalii/go-company-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ── Configuration ────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── Logger ───────────────────────────────────────────────────────────────
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting service", zap.String("name", cfg.App.Name), zap.String("env", cfg.App.Env))

	// ── Database ─────────────────────────────────────────────────────────────
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	log.Info("connected to postgres")

	// ── Kafka ────────────────────────────────────────────────────────────────
	kafkaWriter := infrakafka.NewWriter(cfg.Kafka)
	publisher := eventadapter.NewKafkaPublisher(kafkaWriter)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Error("close kafka writer", zap.Error(err))
		}
	}()

	// ── Repositories ─────────────────────────────────────────────────────────
	companyRepo := repository.NewCompanyRepository(db)
	userRepo := repository.NewUserRepository(db)

	// ── JWT ──────────────────────────────────────────────────────────────────
	jwtMgr := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// ── Use Cases ────────────────────────────────────────────────────────────
	companyUC := company.New(companyRepo, publisher, log)
	authUC := auth.New(userRepo, jwtMgr, log)

	// ── HTTP Server ──────────────────────────────────────────────────────────
	e := server.New(log, cfg.App.Debug)
	e.Use(httpmw.RequestLogger(log))

	companyH := handler.NewCompanyHandler(companyUC)
	authH := handler.NewAuthHandler(authUC)
	router.Register(e, jwtMgr, authH, companyH)

	// ── Graceful Shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + cfg.App.Port
		log.Info("HTTP server listening", zap.String("addr", addr))
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info("server stopped")
	return nil
}
