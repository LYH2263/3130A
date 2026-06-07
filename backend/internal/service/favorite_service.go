package service

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

type FavoriteService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewFavoriteService(db *gorm.DB, log *slog.Logger) *FavoriteService {
	return &FavoriteService{db: db, log: log}
}

func (s *FavoriteService) ToggleFavorite(userID uint, questionID uint) (bool, error) {
	var existing models.Favorite
	result := s.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&existing)

	if result.Error == nil {
		if err := s.db.Delete(&existing).Error; err != nil {
			return false, fmt.Errorf("unfavorite: %w", err)
		}
		s.log.Info("question unfavorited", "userId", userID, "questionId", questionID)
		return false, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("check favorite: %w", result.Error)
	}

	favorite := models.Favorite{
		UserID:     userID,
		QuestionID: questionID,
	}
	if err := s.db.Create(&favorite).Error; err != nil {
		return false, fmt.Errorf("add favorite: %w", err)
	}
	s.log.Info("question favorited", "userId", userID, "questionId", questionID)
	return true, nil
}

func (s *FavoriteService) IsFavorited(userID uint, questionID uint) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Favorite{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check favorite: %w", err)
	}
	return count > 0, nil
}

func (s *FavoriteService) GetFavoriteStatus(userID uint, questionIDs []uint) (map[uint]bool, error) {
	if len(questionIDs) == 0 {
		return map[uint]bool{}, nil
	}

	var favorites []models.Favorite
	if err := s.db.Where("user_id = ? AND question_id IN ?", userID, questionIDs).
		Find(&favorites).Error; err != nil {
		return nil, fmt.Errorf("get favorite status: %w", err)
	}

	status := make(map[uint]bool)
	for _, id := range questionIDs {
		status[id] = false
	}
	for _, f := range favorites {
		status[f.QuestionID] = true
	}
	return status, nil
}

func (s *FavoriteService) GetFavorites(userID uint) ([]dto.FavoriteItem, error) {
	var favorites []models.Favorite
	if err := s.db.Preload("Question").
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&favorites).Error; err != nil {
		return nil, fmt.Errorf("get favorites: %w", err)
	}

	items := make([]dto.FavoriteItem, 0, len(favorites))
	for _, f := range favorites {
		title := ""
		qType := ""
		if f.Question != nil {
			title = f.Question.Title
			qType = f.Question.Type
		}
		items = append(items, dto.FavoriteItem{
			ID:         f.ID,
			QuestionID: f.QuestionID,
			Title:      title,
			Type:       qType,
			CreatedAt:  f.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, nil
}

func (s *FavoriteService) GetFavoriteQuestions(userID uint) ([]models.Question, error) {
	var favorites []models.Favorite
	if err := s.db.Preload("Question.Options").
		Preload("Question.BlankAnswers").
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&favorites).Error; err != nil {
		return nil, fmt.Errorf("get favorite questions: %w", err)
	}

	questions := make([]models.Question, 0, len(favorites))
	for _, f := range favorites {
		if f.Question != nil {
			questions = append(questions, *f.Question)
		}
	}
	return questions, nil
}
