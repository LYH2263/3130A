package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"label3130/backend/internal/auth"
	"label3130/backend/internal/config"
	"label3130/backend/internal/database"
	"label3130/backend/internal/handler"
	"label3130/backend/internal/logger"
	"label3130/backend/internal/seed"
	"label3130/backend/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	db, err := connectWithRetry(cfg.DSN(), 15, 2*time.Second, log)
	if err != nil {
		log.Error("database connect failed", "error", err.Error())
		return
	}

	if err := seed.Run(db, log); err != nil {
		log.Error("database seed failed", "error", err.Error())
		return
	}

	tokens := auth.NewTokenManager(cfg.JWTSecret)
	authSvc := service.NewAuthService(db, tokens, log)
	categorySvc := service.NewCategoryService(db, log)
	tagSvc := service.NewTagService(db, log)
	knowledgePointSvc := service.NewKnowledgePointService(db, log)
	questionSvc := service.NewQuestionService(db, log)
	attemptSvc := service.NewAttemptService(db, log)
	mistakeReviewSvc := service.NewMistakeReviewService(db, log)
	favoriteSvc := service.NewFavoriteService(db, log)
	questionSetSvc := service.NewQuestionSetService(db, log)

	h := handler.New(authSvc, categorySvc, tagSvc, knowledgePointSvc, questionSvc, attemptSvc, mistakeReviewSvc, favoriteSvc, questionSetSvc, tokens, log)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Info("server started", "port", cfg.Port, "env", cfg.AppEnv)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server stopped", "error", err.Error())
	}
}

func connectWithRetry(dsn string, retries int, interval time.Duration, log *slog.Logger) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		db, err := database.Connect(dsn, log)
		if err == nil {
			return db, nil
		}
		lastErr = err
		time.Sleep(interval)
	}
	return nil, fmt.Errorf("connect with retry failed: %w", lastErr)
}
