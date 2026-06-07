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
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Options     []StudentOption `json:"options,omitempty"`
}

func NewQuestionService(db *gorm.DB, log *slog.Logger) *QuestionService {
	return &QuestionService{db: db, log: log}
}

func (s *QuestionService) ListQuestions() ([]models.Question, error) {
	var questions []models.Question
	if err := s.db.Preload("Options").Preload("BlankAnswers").Preload("Tags").Preload("KnowledgePoints").Preload("Explanation").Preload("Category").Order("id desc").Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	return questions, nil
}

func (s *QuestionService) GetQuestion(id uint) (*models.Question, error) {
	var question models.Question
	if err := s.db.Preload("Options").Preload("BlankAnswers").Preload("Tags").Preload("KnowledgePoints").Preload("Explanation").Preload("Category").First(&question, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("get question: %w", err)
	}
	return &question, nil
}

func (s *QuestionService) QueryQuestions(query dto.QuestionQuery, categorySvc *CategoryService) (*dto.PaginatedQuestions, error) {
	db := s.db.Model(&models.Question{})

	if query.Keyword != "" {
		keyword := "%" + strings.ReplaceAll(query.Keyword, "%", "\\%") + "%"
		db = db.Where("questions.title LIKE ? OR questions.description LIKE ?", keyword, keyword)
	}

	if query.CategoryID != nil {
		categoryIDs, err := categorySvc.GetDescendantIDs(*query.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("get category descendants: %w", err)
		}
		db = db.Where("questions.category_id IN ?", categoryIDs)
	}

	if len(query.TagIDs) > 0 {
		subQuery := s.db.Table("question_tags").
			Select("question_id").
			Where("tag_id IN ?", query.TagIDs).
			Group("question_id")

		tagMode := query.TagMode
		if tagMode == "" {
			tagMode = "or"
		}
		if tagMode == "and" {
			subQuery = subQuery.Having("COUNT(DISTINCT tag_id) = ?", len(query.TagIDs))
		}

		db = db.Where("questions.id IN (?)", subQuery)
	}

	if query.CreatedBy != nil {
		db = db.Where("questions.created_by = ?", *query.CreatedBy)
	}

	if query.CreatedFrom != "" {
		if t, err := time.Parse("2006-01-02", query.CreatedFrom); err == nil {
			db = db.Where("questions.created_at >= ?", t)
		}
	}

	if query.CreatedTo != "" {
		if t, err := time.Parse("2006-01-02", query.CreatedTo); err == nil {
			endOfDay := t.Add(24 * time.Hour).Add(-time.Second)
			db = db.Where("questions.created_at <= ?", endOfDay)
		}
	}

	wrongCountSubQuery := s.db.Table("attempt_answers").
		Select("COUNT(*)").
		Where("attempt_answers.question_id = questions.id AND attempt_answers.is_correct = ?", false)

	if query.HasAnswerError != nil {
		if *query.HasAnswerError {
			db = db.Where(`(
				(questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id 
					AND question_options.is_correct = ?
				) != 1)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id 
					AND question_options.is_correct = ?
				) < 2)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM blank_answers 
					WHERE blank_answers.question_id = questions.id
				) < 1)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id 
					AND question_options.is_correct = ?
				) != 1)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id
				) != 2)
			)`, models.QuestionTypeSingle, true, models.QuestionTypeMultiple, true, models.QuestionTypeBlank, models.QuestionTypeJudge, true, models.QuestionTypeJudge)
		} else {
			db = db.Where(`(
				(questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id 
					AND question_options.is_correct = ?
				) = 1)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id 
					AND question_options.is_correct = ?
				) >= 2)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM blank_answers 
					WHERE blank_answers.question_id = questions.id
				) >= 1)
				OR (questions.type = ? AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id 
					AND question_options.is_correct = ?
				) = 1 AND (
					SELECT COUNT(*) FROM question_options 
					WHERE question_options.question_id = questions.id
				) = 2)
			)`, models.QuestionTypeSingle, true, models.QuestionTypeMultiple, true, models.QuestionTypeBlank, models.QuestionTypeJudge, true)
		}
	}

	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "id"
	}
	sortOrder := query.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	var orderClause string
	switch sortBy {
	case "created_at":
		orderClause = "questions.created_at " + sortOrder
	case "wrong_count":
		orderClause = "wrong_count " + sortOrder + ", questions.id desc"
	default:
		orderClause = "questions.id " + sortOrder
	}

	type questionWithWrongCount struct {
		models.Question
		WrongCount int64 `gorm:"column:wrong_count"`
	}

	var rows []questionWithWrongCount
	if err := db.Session(&gorm.Session{}).
		Select("questions.*, (?) as wrong_count", wrongCountSubQuery).
		Order(orderClause).
		Limit(pageSize).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}

	orderedIDs := make([]uint, 0, len(rows))
	wrongCountMap := make(map[uint]int64, len(rows))
	for _, r := range rows {
		orderedIDs = append(orderedIDs, r.ID)
		wrongCountMap[r.ID] = r.WrongCount
	}

	// Load associations on the real Question model. Preloading the many2many
	// relations (Tags/KnowledgePoints) directly on the wrapper struct above
	// makes GORM derive a wrong join column name (question_with_wrong_count_id)
	// and fail, so association loading is done here on models.Question.
	questionMap := make(map[uint]models.Question, len(rows))
	if len(orderedIDs) > 0 {
		var fullQuestions []models.Question
		if err := s.db.
			Preload("Category").Preload("Tags").Preload("KnowledgePoints").Preload("Explanation").
			Preload("Options").Preload("BlankAnswers").
			Where("id IN ?", orderedIDs).Find(&fullQuestions).Error; err != nil {
			return nil, fmt.Errorf("load question associations: %w", err)
		}
		for _, q := range fullQuestions {
			questionMap[q.ID] = q
		}
	}

	creatorIDs := make([]uint, 0, len(rows))
	for _, id := range orderedIDs {
		if q, ok := questionMap[id]; ok && q.CreatedBy > 0 {
			creatorIDs = append(creatorIDs, q.CreatedBy)
		}
	}

	creatorMap := make(map[uint]string)
	if len(creatorIDs) > 0 {
		var users []models.User
		if err := s.db.Where("id IN ?", creatorIDs).Find(&users).Error; err == nil {
			for _, u := range users {
				creatorMap[u.ID] = u.Username
			}
		}
	}

	items := make([]dto.QuestionDetail, 0, len(rows))
	for _, id := range orderedIDs {
		q, ok := questionMap[id]
		if !ok {
			continue
		}
		categoryName := ""
		if q.Category != nil {
			categoryName = q.Category.Name
		}
		expContent := ""
		expRefs := ""
		if q.Explanation != nil {
			expContent = q.Explanation.Content
			expRefs = q.Explanation.References
		}

		hasError := false
		switch q.Type {
		case models.QuestionTypeSingle:
			correctCount := 0
			for _, opt := range q.Options {
				if opt.IsCorrect {
					correctCount++
				}
			}
			hasError = correctCount != 1
		case models.QuestionTypeMultiple:
			correctCount := 0
			for _, opt := range q.Options {
				if opt.IsCorrect {
					correctCount++
				}
			}
			hasError = correctCount < 2
		case models.QuestionTypeJudge:
			correctCount := 0
			for _, opt := range q.Options {
				if opt.IsCorrect {
					correctCount++
				}
			}
			hasError = correctCount != 1 || len(q.Options) != 2
		case models.QuestionTypeBlank:
			hasError = len(q.BlankAnswers) < 1
		}

		items = append(items, dto.QuestionDetail{
			ID:                 q.ID,
			Type:               q.Type,
			Title:              q.Title,
			Description:        q.Description,
			CategoryID:         q.CategoryID,
			CategoryName:       categoryName,
			CreatedBy:          q.CreatedBy,
			CreatedByName:      creatorMap[q.CreatedBy],
			MultipleScore:      q.MultipleScore,
			ExplanationContent: expContent,
			ExplanationRefs:    expRefs,
			KnowledgePoints:    toKnowledgePointInfos(q.KnowledgePoints),
			WrongCount:         wrongCountMap[id],
			HasAnswerError:     hasError,
			CreatedAt:          q.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:          q.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &dto.PaginatedQuestions{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *QuestionService) CreateQuestion(input dto.QuestionInput, createdBy uint, tagSvc *TagService, kpSvc *KnowledgePointService) (*models.Question, error) {
	if err := validateQuestionInput(input); err != nil {
		return nil, err
	}

	question := models.Question{
		Type:          strings.TrimSpace(input.Type),
		Title:         strings.TrimSpace(input.Title),
		Description:   strings.TrimSpace(input.Description),
		CategoryID:    input.CategoryID,
		CreatedBy:     createdBy,
		MultipleScore: strings.TrimSpace(input.MultipleScore),
	}

	if question.MultipleScore == "" {
		question.MultipleScore = models.MultipleScoringAllOrNothing
	}

	if question.Type == models.QuestionTypeBlank {
		question.BlankAnswers = toBlankAnswerModels(input.BlankAnswers)
	} else {
		question.Options = toOptionModels(input.Options)
	}

	hasExplanation := strings.TrimSpace(input.ExplanationContent) != "" || strings.TrimSpace(input.ExplanationRefs) != ""

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&question).Error; err != nil {
			return fmt.Errorf("create question: %w", err)
		}

		if len(input.TagNames) > 0 && tagSvc != nil {
			var tags []models.Tag
			for _, name := range input.TagNames {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				tag, err := tagSvc.GetOrCreateTag(name)
				if err != nil {
					return fmt.Errorf("get or create tag: %w", err)
				}
				tags = append(tags, *tag)
			}
			if len(tags) > 0 {
				if err := tx.Model(&question).Association("Tags").Append(tags); err != nil {
					return fmt.Errorf("associate tags: %w", err)
				}
			}
		}

		if len(input.KnowledgePointNames) > 0 && kpSvc != nil {
			var kps []models.KnowledgePoint
			for _, name := range input.KnowledgePointNames {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				kp, err := kpSvc.GetOrCreateKnowledgePoint(name)
				if err != nil {
					return fmt.Errorf("get or create knowledge point: %w", err)
				}
				kps = append(kps, *kp)
			}
			if len(kps) > 0 {
				if err := tx.Model(&question).Association("KnowledgePoints").Append(kps); err != nil {
					return fmt.Errorf("associate knowledge points: %w", err)
				}
			}
		}

		if hasExplanation {
			explanation := models.QuestionExplanation{
				QuestionID: question.ID,
				Content:    strings.TrimSpace(input.ExplanationContent),
				References: strings.TrimSpace(input.ExplanationRefs),
			}
			if err := tx.Create(&explanation).Error; err != nil {
				return fmt.Errorf("create explanation: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err := s.db.Preload("Options").Preload("BlankAnswers").Preload("Tags").Preload("KnowledgePoints").Preload("Explanation").Preload("Category").First(&question, question.ID).Error; err != nil {
		return nil, fmt.Errorf("reload question: %w", err)
	}

	s.log.Info("question created", "questionID", question.ID, "type", question.Type, "createdBy", createdBy)
	return &question, nil
}

func (s *QuestionService) UpdateQuestion(questionID uint, input dto.QuestionInput, tagSvc *TagService, kpSvc *KnowledgePointService) (*models.Question, error) {
	if err := validateQuestionInput(input); err != nil {
		return nil, err
	}

	var question models.Question
	if err := s.db.Preload("Options").Preload("BlankAnswers").Preload("Explanation").First(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("find question: %w", err)
	}

	question.Type = strings.TrimSpace(input.Type)
	question.Title = strings.TrimSpace(input.Title)
	question.Description = strings.TrimSpace(input.Description)
	question.CategoryID = input.CategoryID
	question.MultipleScore = strings.TrimSpace(input.MultipleScore)
	if question.MultipleScore == "" {
		question.MultipleScore = models.MultipleScoringAllOrNothing
	}

	hasExplanation := strings.TrimSpace(input.ExplanationContent) != "" || strings.TrimSpace(input.ExplanationRefs) != ""

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&question).Updates(map[string]any{
			"type":           question.Type,
			"title":          question.Title,
			"description":    question.Description,
			"category_id":    question.CategoryID,
			"multiple_score": question.MultipleScore,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuestionOption{}).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", question.ID).Delete(&models.BlankAnswer{}).Error; err != nil {
			return err
		}

		if question.Type == models.QuestionTypeBlank {
			answers := toBlankAnswerModels(input.BlankAnswers)
			for i := range answers {
				answers[i].QuestionID = question.ID
			}
			if err := tx.Create(&answers).Error; err != nil {
				return err
			}
		} else {
			options := toOptionModels(input.Options)
			for i := range options {
				options[i].QuestionID = question.ID
			}
			if err := tx.Create(&options).Error; err != nil {
				return err
			}
		}

		if tagSvc != nil {
			if err := tx.Model(&question).Association("Tags").Clear(); err != nil {
				return fmt.Errorf("clear tags: %w", err)
			}
			if len(input.TagNames) > 0 {
				var tags []models.Tag
				for _, name := range input.TagNames {
					name = strings.TrimSpace(name)
					if name == "" {
						continue
					}
					tag, err := tagSvc.GetOrCreateTag(name)
					if err != nil {
						return fmt.Errorf("get or create tag: %w", err)
					}
					tags = append(tags, *tag)
				}
				if len(tags) > 0 {
					if err := tx.Model(&question).Association("Tags").Append(tags); err != nil {
						return fmt.Errorf("associate tags: %w", err)
					}
				}
			}
		}

		if kpSvc != nil {
			if err := tx.Model(&question).Association("KnowledgePoints").Clear(); err != nil {
				return fmt.Errorf("clear knowledge points: %w", err)
			}
			if len(input.KnowledgePointNames) > 0 {
				var kps []models.KnowledgePoint
				for _, name := range input.KnowledgePointNames {
					name = strings.TrimSpace(name)
					if name == "" {
						continue
					}
					kp, err := kpSvc.GetOrCreateKnowledgePoint(name)
					if err != nil {
						return fmt.Errorf("get or create knowledge point: %w", err)
					}
					kps = append(kps, *kp)
				}
				if len(kps) > 0 {
					if err := tx.Model(&question).Association("KnowledgePoints").Append(kps); err != nil {
						return fmt.Errorf("associate knowledge points: %w", err)
					}
				}
			}
		}

		if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuestionExplanation{}).Error; err != nil {
			return fmt.Errorf("delete old explanation: %w", err)
		}
		if hasExplanation {
			explanation := models.QuestionExplanation{
				QuestionID: question.ID,
				Content:    strings.TrimSpace(input.ExplanationContent),
				References: strings.TrimSpace(input.ExplanationRefs),
			}
			if err := tx.Create(&explanation).Error; err != nil {
				return fmt.Errorf("create explanation: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}

	if err := s.db.Preload("Options").Preload("BlankAnswers").Preload("Tags").Preload("KnowledgePoints").Preload("Explanation").Preload("Category").First(&question, question.ID).Error; err != nil {
		return nil, fmt.Errorf("reload question: %w", err)
	}
	return &question, nil
}

func (s *QuestionService) DeleteQuestion(questionID uint) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", questionID).Delete(&models.QuestionOption{}).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", questionID).Delete(&models.BlankAnswer{}).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", questionID).Delete(&models.QuestionTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", questionID).Delete(&models.QuestionKnowledgePoint{}).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", questionID).Delete(&models.QuestionExplanation{}).Error; err != nil {
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

func (s *QuestionService) UploadFromJSON(data []byte, createdBy uint, tagSvc *TagService, kpSvc *KnowledgePointService) (int, error) {
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
		if item.Type == "" {
			item.Type = models.QuestionTypeSingle
		}
		if err := validateQuestionInput(item); err != nil {
			s.log.Warn("upload question skipped", "error", err.Error())
			continue
		}
		if _, err := s.CreateQuestion(item, createdBy, tagSvc, kpSvc); err != nil {
			s.log.Warn("upload question failed", "error", err.Error())
			continue
		}
		count++
	}
	return count, nil
}

func (s *QuestionService) GetQuizQuestions(limit int) ([]StudentQuestion, error) {
	var questions []models.Question
	query := s.db.Preload("Options").Preload("BlankAnswers").Order("id asc")
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

func validateQuestionInput(input dto.QuestionInput) error {
	qType := strings.TrimSpace(input.Type)
	if qType == "" {
		return ErrInvalidQuestion
	}

	switch qType {
	case models.QuestionTypeSingle:
		return validateSingleQuestion(input)
	case models.QuestionTypeMultiple:
		return validateMultipleQuestion(input)
	case models.QuestionTypeJudge:
		return validateJudgeQuestion(input)
	case models.QuestionTypeBlank:
		return validateBlankQuestion(input)
	default:
		return ErrInvalidQuestionType
	}
}

func validateSingleQuestion(input dto.QuestionInput) error {
	if len(input.Options) < 2 || len(input.Options) > 6 {
		return ErrInvalidSingleOptionCount
	}
	correctCount := 0
	for _, opt := range input.Options {
		if strings.TrimSpace(opt.Content) == "" {
			return ErrInvalidOptionContent
		}
		if opt.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return ErrInvalidSingleCorrect
	}
	return nil
}

func validateMultipleQuestion(input dto.QuestionInput) error {
	if len(input.Options) < 2 || len(input.Options) > 6 {
		return ErrInvalidMultipleOptionCount
	}
	correctCount := 0
	for _, opt := range input.Options {
		if strings.TrimSpace(opt.Content) == "" {
			return ErrInvalidOptionContent
		}
		if opt.IsCorrect {
			correctCount++
		}
	}
	if correctCount < 2 {
		return ErrInvalidMultipleCorrect
	}
	return nil
}

func validateJudgeQuestion(input dto.QuestionInput) error {
	if len(input.Options) != 2 {
		return ErrInvalidJudgeOptionCount
	}
	for _, opt := range input.Options {
		if strings.TrimSpace(opt.Content) == "" {
			return ErrInvalidOptionContent
		}
	}
	correctCount := 0
	for _, opt := range input.Options {
		if opt.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return ErrInvalidJudgeCorrect
	}
	return nil
}

func validateBlankQuestion(input dto.QuestionInput) error {
	if len(input.BlankAnswers) < 1 {
		return ErrInvalidBlankAnswerCount
	}
	for _, ans := range input.BlankAnswers {
		if strings.TrimSpace(ans.Answer) == "" {
			return ErrInvalidBlankAnswer
		}
	}
	return nil
}

func toOptionModels(inputs []dto.QuestionOptionInput) []models.QuestionOption {
	options := make([]models.QuestionOption, 0, len(inputs))
	for i, opt := range inputs {
		options = append(options, models.QuestionOption{
			Content:   strings.TrimSpace(opt.Content),
			IsCorrect: opt.IsCorrect,
			SortOrder: i,
		})
	}
	return options
}

func toBlankAnswerModels(inputs []dto.BlankAnswerInput) []models.BlankAnswer {
	answers := make([]models.BlankAnswer, 0, len(inputs))
	for i, ans := range inputs {
		matchMode := ans.MatchMode
		if matchMode == "" {
			matchMode = models.BlankMatchExact
		}
		answers = append(answers, models.BlankAnswer{
			Answer:    strings.TrimSpace(ans.Answer),
			MatchMode: matchMode,
			SortOrder: i,
		})
	}
	return answers
}

func (s *QuestionService) GetExplanationForStudent(questionID uint, userID uint) (*dto.QuestionExplanationDTO, error) {
	var answerCount int64
	if err := s.db.Model(&models.AttemptAnswer{}).
		Joins("JOIN attempts ON attempt_answers.attempt_id = attempts.id").
		Where("attempt_answers.question_id = ? AND attempts.user_id = ?", questionID, userID).
		Count(&answerCount).Error; err != nil {
		return nil, fmt.Errorf("check answer existence: %w", err)
	}
	if answerCount == 0 {
		return nil, ErrExplanationUnauthorized
	}

	var question models.Question
	if err := s.db.Preload("Explanation").Preload("KnowledgePoints").
		First(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("load question: %w", err)
	}

	result := &dto.QuestionExplanationDTO{
		KnowledgePoints: toKnowledgePointInfos(question.KnowledgePoints),
	}
	if question.Explanation != nil {
		result.Content = question.Explanation.Content
		result.References = question.Explanation.References
	}

	return result, nil
}

func (s *QuestionService) BatchGetExplanationsForStudent(questionIDs []uint, userID uint) (map[uint]*dto.QuestionExplanationDTO, error) {
	if len(questionIDs) == 0 {
		return map[uint]*dto.QuestionExplanationDTO{}, nil
	}

	answeredQuestionIDs := make(map[uint]bool)
	var answers []models.AttemptAnswer
	if err := s.db.
		Joins("JOIN attempts ON attempt_answers.attempt_id = attempts.id").
		Where("attempt_answers.question_id IN ? AND attempts.user_id = ?", questionIDs, userID).
		Find(&answers).Error; err != nil {
		return nil, fmt.Errorf("check answer existence: %w", err)
	}
	for _, a := range answers {
		answeredQuestionIDs[a.QuestionID] = true
	}

	result := make(map[uint]*dto.QuestionExplanationDTO)
	for _, id := range questionIDs {
		if !answeredQuestionIDs[id] {
			continue
		}
		result[id] = nil
	}

	if len(result) == 0 {
		return result, nil
	}

	ids := make([]uint, 0, len(result))
	for id := range result {
		ids = append(ids, id)
	}

	var questions []models.Question
	if err := s.db.Preload("Explanation").Preload("KnowledgePoints").
		Where("id IN ?", ids).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load questions: %w", err)
	}

	for _, q := range questions {
		dto := &dto.QuestionExplanationDTO{
			KnowledgePoints: toKnowledgePointInfos(q.KnowledgePoints),
		}
		if q.Explanation != nil {
			dto.Content = q.Explanation.Content
			dto.References = q.Explanation.References
		}
		result[q.ID] = dto
	}

	return result, nil
}
