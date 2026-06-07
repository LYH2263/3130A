package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"label3130/backend/internal/models"
)

func Connect(dsn string, log *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("resolve sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := autoMigrate(db, log); err != nil {
		return nil, err
	}

	return db, nil
}

func autoMigrate(db *gorm.DB, log *slog.Logger) error {
	if err := db.AutoMigrate(
		&models.ClassRoom{},
		&models.User{},
		&models.Question{},
		&models.QuestionOption{},
		&models.BlankAnswer{},
		&models.Attempt{},
		&models.AttemptAnswer{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	if err := migrateLegacyData(db, log); err != nil {
		return fmt.Errorf("migrate legacy data: %w", err)
	}

	return nil
}

func migrateLegacyData(db *gorm.DB, log *slog.Logger) error {
	result := db.Model(&models.Question{}).
		Where("type = '' OR type IS NULL").
		Updates(map[string]interface{}{
			"type":            models.QuestionTypeSingle,
			"multiple_score":  models.MultipleScoringAllOrNothing,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		log.Info("migrated legacy questions to single type", "count", result.RowsAffected)
	}

	var count int64
	if err := db.Model(&models.Question{}).Count(&count).Error; err != nil {
		return err
	}
	log.Info("question migration check complete", "totalQuestions", count)

	return nil
}
