package service

import (
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

type QuestionSetService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewQuestionSetService(db *gorm.DB, log *slog.Logger) *QuestionSetService {
	return &QuestionSetService{db: db, log: log}
}

var (
	ErrQuestionSetNotFound = errors.New("question set not found")
	ErrQuestionSetNameEmpty = errors.New("question set name cannot be empty")
	ErrQuestionSetForbidden = errors.New("access to question set forbidden")
)

func (s *QuestionSetService) ListSets(userID uint) ([]dto.QuestionSetDTO, error) {
	var sets []models.QuestionSet
	if err := s.db.Where("user_id = ?", userID).
		Order("sort_order asc, id desc").
		Find(&sets).Error; err != nil {
		return nil, fmt.Errorf("list question sets: %w", err)
	}

	result := make([]dto.QuestionSetDTO, 0, len(sets))
	for _, set := range sets {
		qCount := 0
		if set.QuestionIDs != nil {
			qCount = len(set.QuestionIDs)
		}
		result = append(result, dto.QuestionSetDTO{
			ID:            set.ID,
			Name:          set.Name,
			Description:   set.Description,
			QuestionIDs:   set.QuestionIDs,
			QuestionCount: qCount,
			SortOrder:     set.SortOrder,
			CreatedAt:     set.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     set.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *QuestionSetService) GetSet(userID uint, setID uint) (*dto.QuestionSetDTO, error) {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrQuestionSetNotFound
		}
		return nil, fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return nil, ErrQuestionSetForbidden
	}

	qCount := 0
	if set.QuestionIDs != nil {
		qCount = len(set.QuestionIDs)
	}

	return &dto.QuestionSetDTO{
		ID:            set.ID,
		Name:          set.Name,
		Description:   set.Description,
		QuestionIDs:   set.QuestionIDs,
		QuestionCount: qCount,
		SortOrder:     set.SortOrder,
		CreatedAt:     set.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     set.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *QuestionSetService) CreateSet(userID uint, req dto.CreateQuestionSetRequest) (*dto.QuestionSetDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrQuestionSetNameEmpty
	}

	set := models.QuestionSet{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		QuestionIDs: req.QuestionIDs,
	}

	if err := s.db.Create(&set).Error; err != nil {
		return nil, fmt.Errorf("create question set: %w", err)
	}

	s.log.Info("question set created", "setId", set.ID, "userId", userID)

	qCount := 0
	if set.QuestionIDs != nil {
		qCount = len(set.QuestionIDs)
	}

	return &dto.QuestionSetDTO{
		ID:            set.ID,
		Name:          set.Name,
		Description:   set.Description,
		QuestionIDs:   set.QuestionIDs,
		QuestionCount: qCount,
		SortOrder:     set.SortOrder,
		CreatedAt:     set.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     set.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *QuestionSetService) UpdateSet(userID uint, setID uint, req dto.UpdateQuestionSetRequest) (*dto.QuestionSetDTO, error) {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrQuestionSetNotFound
		}
		return nil, fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return nil, ErrQuestionSetForbidden
	}

	updated := false
	if req.Name != "" {
		set.Name = strings.TrimSpace(req.Name)
		updated = true
	}
	if req.Description != "" {
		set.Description = strings.TrimSpace(req.Description)
		updated = true
	}
	if req.QuestionIDs != nil {
		set.QuestionIDs = req.QuestionIDs
		updated = true
	}
	if req.SortOrder != nil {
		set.SortOrder = *req.SortOrder
		updated = true
	}

	if updated {
		if err := s.db.Save(&set).Error; err != nil {
			return nil, fmt.Errorf("update question set: %w", err)
		}
		s.log.Info("question set updated", "setId", set.ID, "userId", userID)
	}

	qCount := 0
	if set.QuestionIDs != nil {
		qCount = len(set.QuestionIDs)
	}

	return &dto.QuestionSetDTO{
		ID:            set.ID,
		Name:          set.Name,
		Description:   set.Description,
		QuestionIDs:   set.QuestionIDs,
		QuestionCount: qCount,
		SortOrder:     set.SortOrder,
		CreatedAt:     set.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     set.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *QuestionSetService) DeleteSet(userID uint, setID uint) error {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrQuestionSetNotFound
		}
		return fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return ErrQuestionSetForbidden
	}

	if err := s.db.Delete(&set).Error; err != nil {
		return fmt.Errorf("delete question set: %w", err)
	}

	s.log.Info("question set deleted", "setId", setID, "userId", userID)
	return nil
}

func (s *QuestionSetService) AddQuestions(userID uint, setID uint, questionIDs []uint) error {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrQuestionSetNotFound
		}
		return fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return ErrQuestionSetForbidden
	}

	existing := make(map[uint]bool)
	if set.QuestionIDs != nil {
		for _, id := range set.QuestionIDs {
			existing[id] = true
		}
	}

	newIDs := make([]uint, 0, len(set.QuestionIDs)+len(questionIDs))
	if set.QuestionIDs != nil {
		newIDs = append(newIDs, set.QuestionIDs...)
	}
	for _, id := range questionIDs {
		if !existing[id] {
			newIDs = append(newIDs, id)
			existing[id] = true
		}
	}

	set.QuestionIDs = newIDs
	if err := s.db.Save(&set).Error; err != nil {
		return fmt.Errorf("add questions to set: %w", err)
	}

	s.log.Info("questions added to set", "setId", setID, "userId", userID, "count", len(questionIDs))
	return nil
}

func (s *QuestionSetService) RemoveQuestion(userID uint, setID uint, questionID uint) error {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrQuestionSetNotFound
		}
		return fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return ErrQuestionSetForbidden
	}

	newIDs := make([]uint, 0, len(set.QuestionIDs))
	for _, id := range set.QuestionIDs {
		if id != questionID {
			newIDs = append(newIDs, id)
		}
	}

	set.QuestionIDs = newIDs
	if err := s.db.Save(&set).Error; err != nil {
		return fmt.Errorf("remove question from set: %w", err)
	}

	s.log.Info("question removed from set", "setId", setID, "userId", userID, "questionId", questionID)
	return nil
}

func (s *QuestionSetService) ReorderQuestions(userID uint, setID uint, questionIDs []uint) error {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrQuestionSetNotFound
		}
		return fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return ErrQuestionSetForbidden
	}

	set.QuestionIDs = questionIDs
	if err := s.db.Save(&set).Error; err != nil {
		return fmt.Errorf("reorder questions in set: %w", err)
	}

	s.log.Info("questions reordered in set", "setId", setID, "userId", userID)
	return nil
}

func (s *QuestionSetService) GetSetQuestions(userID uint, setID uint) ([]StudentQuestion, error) {
	var set models.QuestionSet
	if err := s.db.First(&set, setID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrQuestionSetNotFound
		}
		return nil, fmt.Errorf("get question set: %w", err)
	}

	if set.UserID != userID {
		return nil, ErrQuestionSetForbidden
	}

	if len(set.QuestionIDs) == 0 {
		return []StudentQuestion{}, nil
	}

	var questions []models.Question
	qids := make([]uint, len(set.QuestionIDs))
	copy(qids, set.QuestionIDs)
	if err := s.db.Preload("Options").Preload("BlankAnswers").
		Where("id IN ?", qids).
		Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("get set questions: %w", err)
	}

	qMap := make(map[uint]models.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]StudentQuestion, 0, len(set.QuestionIDs))

	for _, qid := range set.QuestionIDs {
		q, ok := qMap[qid]
		if !ok {
			continue
		}

		sq := StudentQuestion{
			ID:          q.ID,
			Type:        q.Type,
			Title:       q.Title,
			Description: q.Description,
		}

		if q.Type != models.QuestionTypeBlank {
			opts := make([]StudentOption, 0, len(q.Options))
			for _, opt := range q.Options {
				opts = append(opts, StudentOption{ID: opt.ID, Content: opt.Content})
			}
			r.Shuffle(len(opts), func(i, j int) {
				opts[i], opts[j] = opts[j], opts[i]
			})
			sq.Options = opts
		}

		result = append(result, sq)
	}

	return result, nil
}

func (s *QuestionSetService) GetSetDetailWithQuestions(userID uint, setID uint) (*dto.QuestionSetDTO, []StudentQuestion, error) {
	set, err := s.GetSet(userID, setID)
	if err != nil {
		return nil, nil, err
	}

	questions, err := s.GetSetQuestions(userID, setID)
	if err != nil {
		return nil, nil, err
	}

	return set, questions, nil
}
