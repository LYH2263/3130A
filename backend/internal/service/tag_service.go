package service

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"label3130/backend/internal/models"
)

type TagService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewTagService(db *gorm.DB, log *slog.Logger) *TagService {
	return &TagService{db: db, log: log}
}

func (s *TagService) ListTags() ([]models.Tag, error) {
	var tags []models.Tag
	if err := s.db.Order("id asc").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return tags, nil
}

func (s *TagService) GetTag(id uint) (*models.Tag, error) {
	var tag models.Tag
	if err := s.db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("get tag: %w", err)
	}
	return &tag, nil
}

func (s *TagService) GetOrCreateTag(name string) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTagNameEmpty
	}

	var tag models.Tag
	err := s.db.Where("name = ?", name).First(&tag).Error
	if err == nil {
		return &tag, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find tag: %w", err)
	}

	tag = models.Tag{Name: name}
	if err := s.db.Create(&tag).Error; err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	s.log.Info("tag created", "tagID", tag.ID, "name", tag.Name)
	return &tag, nil
}

func (s *TagService) CreateTag(name string) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTagNameEmpty
	}

	var count int64
	s.db.Model(&models.Tag{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return nil, ErrTagExists
	}

	tag := models.Tag{Name: name}
	if err := s.db.Create(&tag).Error; err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	s.log.Info("tag created", "tagID", tag.ID, "name", tag.Name)
	return &tag, nil
}

func (s *TagService) UpdateTag(id uint, name string) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTagNameEmpty
	}

	var tag models.Tag
	if err := s.db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("find tag: %w", err)
	}

	var existing models.Tag
	err := s.db.Where("name = ? AND id != ?", name, id).First(&existing).Error
	if err == nil {
		return nil, ErrTagExists
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("check tag: %w", err)
	}

	tag.Name = name
	if err := s.db.Save(&tag).Error; err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}

	s.log.Info("tag updated", "tagID", tag.ID)
	return &tag, nil
}

func (s *TagService) DeleteTag(id uint) error {
	result := s.db.Delete(&models.Tag{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	s.log.Info("tag deleted", "tagID", id)
	return nil
}

func (s *TagService) GetQuestionTags(questionID uint) ([]models.Tag, error) {
	var question models.Question
	if err := s.db.Preload("Tags").First(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("get question: %w", err)
	}
	return question.Tags, nil
}
