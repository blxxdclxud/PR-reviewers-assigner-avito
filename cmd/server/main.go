package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/postgres"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/config"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting PR manager service")

	// Load config
	cfg := config.LoadConfig()

	// Connect to postgres database
	logger.Info("connecting to database",
		zap.String("host", cfg.DBConfig.Host),
		zap.Int("port", cfg.DBConfig.Port),
	)
	db, err := postgres.NewDB(cfg.DBConfig)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer postgres.CloseDB(db)

	logger.Info("database connection established")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db, logger)
	teamRepo := postgres.NewTeamRepository(db, logger)
	prRepo := postgres.NewPullRequestRepository(db, logger)

	statsRepo := postgres.NewStatsRepository(db)

	// Initialize use cases
	userUC := usecase.NewUserUseCase(userRepo, prRepo, teamRepo, db)
	teamUC := usecase.NewTeamUseCase(teamRepo, userRepo, db)
	prUC := usecase.NewPRUseCase(userRepo, prRepo, db)

	statsUC := usecase.NewStatsUseCase(statsRepo)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	logger.Info("starting HTTP server",
		zap.String("addr", addr))

	server := &http.Server{
		Addr:    addr,
		Handler: httpAdapter.SetupRouter(teamUC, userUC, prUC, statsUC),
	}

	// Run server
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for shutdown signal to stop server gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received")

	// Shutdown servers
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
