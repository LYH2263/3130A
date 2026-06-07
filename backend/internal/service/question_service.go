package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"

	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

type QuestionService struct {
	db  *gorm.DB
	log *slog.Logger
}

type StudentOption struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
}

type StudentQuestion struct {
	ID          uint            `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Options     []StudentOption `json:"options"`
}

func NewQuestionService(db *gorm.DB, log *slog.Logger) *QuestionService {
	return &QuestionService{db: db, log: log}
}

func (s *QuestionService) ListQuestions() ([]models.Question, error) {
	var questions []models.Question
	if err := s.db.Preload("Options").Order("id desc").Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	return questions, nil
}

func (s *QuestionService) CreateQuestion(input dto.QuestionInput, createdBy uint) (*models.Question, error) {
	if !isQuestionValid(input.Options) {
		return nil, ErrInvalidQuestion
	}

	question := models.Question{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		CreatedBy:   createdBy,
		Options:     toOptionModels(input.Options),
	}

	if err := s.db.Create(&question).Error; err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}

	if err := s.db.Preload("Options").First(&question, question.ID).Error; err != nil {
		return nil, fmt.Errorf("reload question: %w", err)
	}

	s.log.Info("question created", "questionID", question.ID, "createdBy", createdBy)
	return &question, nil
}

func (s *QuestionService) UpdateQuestion(questionID uint, input dto.QuestionInput) (*models.Question, error) {
	if !isQuestionValid(input.Options) {
		return nil, ErrInvalidQuestion
	}

	var question models.Question
	if err := s.db.Preload("Options").First(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("find question: %w", err)
	}

	question.Title = strings.TrimSpace(input.Title)
	question.Description = strings.TrimSpace(input.Description)

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&question).Updates(map[string]any{
			"title":       question.Title,
			"description": question.Description,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuestionOption{}).Error; err != nil {
			return err
		}
		options := toOptionModels(input.Options)
		for i := range options {
			options[i].QuestionID = question.ID
		}
		if err := tx.Create(&options).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}

	if err := s.db.Preload("Options").First(&question, question.ID).Error; err != nil {
		return nil, fmt.Errorf("reload question: %w", err)
	}
	return &question, nil
}

func (s *QuestionService) DeleteQuestion(questionID uint) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", questionID).Delete(&models.QuestionOption{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&models.Question{}, questionID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrQuestionNotFound
		}
		return nil
	}); err != nil {
		return err
	}
	s.log.Info("question deleted", "questionID", questionID)
	return nil
}

func (s *QuestionService) UploadFromJSON(data []byte, createdBy uint) (int, error) {
	var payload dto.UploadQuestionPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		var arrayPayload []dto.QuestionInput
		if errArray := json.Unmarshal(data, &arrayPayload); errArray != nil {
			return 0, fmt.Errorf("invalid upload payload")
		}
		payload.Questions = arrayPayload
	}

	if len(payload.Questions) == 0 {
		return 0, fmt.Errorf("upload payload is empty")
	}

	count := 0
	for _, item := range payload.Questions {
		if !isQuestionValid(item.Options) {
			continue
		}
		if _, err := s.CreateQuestion(item, createdBy); err != nil {
			s.log.Warn("upload question failed", "error", err.Error())
			continue
		}
		count++
	}
	return count, nil
}

func (s *QuestionService) GetQuizQuestions(limit int) ([]StudentQuestion, error) {
	var questions []models.Question
	query := s.db.Preload("Options").Order("id asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load questions: %w", err)
	}
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]StudentQuestion, 0, len(questions))
	for _, q := range questions {
		opts := make([]StudentOption, 0, len(q.Options))
		for _, opt := range q.Options {
			opts = append(opts, StudentOption{ID: opt.ID, Content: opt.Content})
		}
		r.Shuffle(len(opts), func(i, j int) {
			opts[i], opts[j] = opts[j], opts[i]
		})
		result = append(result, StudentQuestion{
			ID:          q.ID,
			Title:       q.Title,
			Description: q.Description,
			Options:     opts,
		})
	}
	return result, nil
}

func isQuestionValid(options []dto.QuestionOptionInput) bool {
	if len(options) < 2 {
		return false
	}
	correctCount := 0
	for _, opt := range options {
		if strings.TrimSpace(opt.Content) == "" {
			return false
		}
		if opt.IsCorrect {
			correctCount++
		}
	}
	return correctCount == 1
}

func toOptionModels(inputs []dto.QuestionOptionInput) []models.QuestionOption {
	options := make([]models.QuestionOption, 0, len(inputs))
	for _, opt := range inputs {
		options = append(options, models.QuestionOption{
			Content:   strings.TrimSpace(opt.Content),
			IsCorrect: opt.IsCorrect,
		})
	}
	return options
}
