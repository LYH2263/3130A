package service

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"label3130/backend/internal/models"
)

type ExamConfigService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewExamConfigService(db *gorm.DB, log *slog.Logger) *ExamConfigService {
	return &ExamConfigService{db: db, log: log}
}

func (s *ExamConfigService) List() ([]models.ExamConfig, error) {
	var configs []models.ExamConfig
	if err := s.db.Order("is_default desc, id asc").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("list exam configs: %w", err)
	}
	return configs, nil
}

func (s *ExamConfigService) GetByID(id uint) (*models.ExamConfig, error) {
	var config models.ExamConfig
	if err := s.db.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrExamConfigNotFound
		}
		return nil, fmt.Errorf("get exam config: %w", err)
	}
	return &config, nil
}

func (s *ExamConfigService) GetDefault() (*models.ExamConfig, error) {
	var config models.ExamConfig
	if err := s.db.Where("is_default = ?", true).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrExamConfigNotFound
		}
		return nil, fmt.Errorf("get default exam config: %w", err)
	}
	return &config, nil
}

func (s *ExamConfigService) Create(name string, durationMinutes int, allowEarlySubmit, forceSubmitOnTimeout, isDefault bool) (*models.ExamConfig, error) {
	if name == "" {
		return nil, ErrExamConfigNameEmpty
	}
	if durationMinutes <= 0 {
		return nil, ErrInvalidDuration
	}

	if isDefault {
		var count int64
		if err := s.db.Model(&models.ExamConfig{}).Where("is_default = ?", true).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("check default config: %w", err)
		}
		if count > 0 {
			return nil, ErrDefaultExamConfigExists
		}
	}

	config := models.ExamConfig{
		Name:                 name,
		DurationMinutes:      durationMinutes,
		AllowEarlySubmit:     allowEarlySubmit,
		ForceSubmitOnTimeout: forceSubmitOnTimeout,
		IsDefault:            isDefault,
	}
	if err := s.db.Create(&config).Error; err != nil {
		return nil, fmt.Errorf("create exam config: %w", err)
	}
	return &config, nil
}

func (s *ExamConfigService) Update(id uint, name string, durationMinutes int, allowEarlySubmit, forceSubmitOnTimeout, isDefault bool) (*models.ExamConfig, error) {
	config, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, ErrExamConfigNameEmpty
	}
	if durationMinutes <= 0 {
		return nil, ErrInvalidDuration
	}

	if isDefault && !config.IsDefault {
		var count int64
		if err := s.db.Model(&models.ExamConfig{}).Where("is_default = ? AND id != ?", true, id).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("check default config: %w", err)
		}
		if count > 0 {
			return nil, ErrDefaultExamConfigExists
		}
	}

	config.Name = name
	config.DurationMinutes = durationMinutes
	config.AllowEarlySubmit = allowEarlySubmit
	config.ForceSubmitOnTimeout = forceSubmitOnTimeout
	config.IsDefault = isDefault

	if err := s.db.Save(config).Error; err != nil {
		return nil, fmt.Errorf("update exam config: %w", err)
	}
	return config, nil
}

func (s *ExamConfigService) Delete(id uint) error {
	result := s.db.Delete(&models.ExamConfig{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete exam config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrExamConfigNotFound
	}
	return nil
}

func (s *ExamConfigService) CalculateDeadline(startedAt time.Time, durationMinutes int) time.Time {
	return startedAt.Add(time.Duration(durationMinutes) * time.Minute)
}

func (s *ExamConfigService) GetOrCreateDefault() (*models.ExamConfig, error) {
	config, err := s.GetDefault()
	if err == nil {
		return config, nil
	}
	if err != ErrExamConfigNotFound {
		return nil, err
	}

	return s.Create("标准测试", 30, true, false, true)
}
