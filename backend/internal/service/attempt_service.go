package service

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"gorm.io/gorm"

	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

type AttemptService struct {
	db  *gorm.DB
	log *slog.Logger
}

type SubmitResult struct {
	AttemptID uint   `json:"attemptId"`
	Score     int    `json:"score"`
	Total     int    `json:"total"`
	Rate      string `json:"rate"`
}

type StudentMistake struct {
	QuestionID    uint   `json:"questionId"`
	Title         string `json:"title"`
	WrongCount    int64  `json:"wrongCount"`
	CorrectOption string `json:"correctOption"`
}

type ClassWrongStat struct {
	ClassID    uint   `json:"classId"`
	ClassName  string `json:"className"`
	QuestionID uint   `json:"questionId"`
	Question   string `json:"question"`
	WrongCount int64  `json:"wrongCount"`
}

type RecentAttempt struct {
	ID        uint   `json:"id"`
	Student   string `json:"student"`
	ClassName string `json:"className"`
	Score     int    `json:"score"`
	Total     int    `json:"total"`
	CreatedAt string `json:"createdAt"`
}

type Overview struct {
	StudentCount  int64 `json:"studentCount"`
	ClassCount    int64 `json:"classCount"`
	QuestionCount int64 `json:"questionCount"`
	AttemptCount  int64 `json:"attemptCount"`
}

func NewAttemptService(db *gorm.DB, log *slog.Logger) *AttemptService {
	return &AttemptService{db: db, log: log}
}

func (s *AttemptService) Submit(userID uint, classID uint, req dto.SubmitRequest) (*SubmitResult, error) {
	if len(req.Answers) == 0 {
		return nil, ErrInvalidSubmission
	}

	questionIDs := uniqueQuestionIDs(req.Answers)
	var questions []models.Question
	if err := s.db.Preload("Options").Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load questions: %w", err)
	}
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	questionMap := make(map[uint]models.Question, len(questions))
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	answersModel := make([]models.AttemptAnswer, 0, len(req.Answers))
	score := 0

	for _, answer := range req.Answers {
		question, ok := questionMap[answer.QuestionID]
		if !ok {
			return nil, ErrInvalidSubmission
		}

		selectedValid := false
		correct := false
		for _, opt := range question.Options {
			if opt.ID == answer.OptionID {
				selectedValid = true
				if opt.IsCorrect {
					correct = true
				}
				break
			}
		}
		if !selectedValid {
			return nil, ErrInvalidSubmission
		}
		if correct {
			score++
		}

		answersModel = append(answersModel, models.AttemptAnswer{
			QuestionID:       answer.QuestionID,
			SelectedOptionID: answer.OptionID,
			IsCorrect:        correct,
		})
	}

	total := len(req.Answers)
	attempt := models.Attempt{
		UserID:  userID,
		ClassID: classID,
		Score:   score,
		Total:   total,
		Answers: answersModel,
	}
	if err := s.db.Create(&attempt).Error; err != nil {
		return nil, fmt.Errorf("save attempt: %w", err)
	}

	rate := fmt.Sprintf("%.0f%%", (float64(score)/float64(total))*100)
	s.log.Info("attempt submitted", "attemptID", attempt.ID, "userID", userID, "score", score, "total", total)

	return &SubmitResult{AttemptID: attempt.ID, Score: score, Total: total, Rate: rate}, nil
}

func (s *AttemptService) StudentAttempts(userID uint) ([]models.Attempt, error) {
	var attempts []models.Attempt
	if err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("load student attempts: %w", err)
	}
	return attempts, nil
}

func (s *AttemptService) StudentMistakes(userID uint) ([]StudentMistake, error) {
	var attempts []models.Attempt
	if err := s.db.Where("user_id = ?", userID).Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("load attempts: %w", err)
	}
	if len(attempts) == 0 {
		return []StudentMistake{}, nil
	}

	attemptIDs := make([]uint, 0, len(attempts))
	for _, a := range attempts {
		attemptIDs = append(attemptIDs, a.ID)
	}

	var wrongAnswers []models.AttemptAnswer
	if err := s.db.Where("attempt_id IN ? AND is_correct = ?", attemptIDs, false).Find(&wrongAnswers).Error; err != nil {
		return nil, fmt.Errorf("load wrong answers: %w", err)
	}
	if len(wrongAnswers) == 0 {
		return []StudentMistake{}, nil
	}

	wrongCountMap := map[uint]int64{}
	for _, item := range wrongAnswers {
		wrongCountMap[item.QuestionID]++
	}

	questionIDs := make([]uint, 0, len(wrongCountMap))
	for id := range wrongCountMap {
		questionIDs = append(questionIDs, id)
	}
	var questions []models.Question
	if err := s.db.Preload("Options").Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load mistake questions: %w", err)
	}

	result := make([]StudentMistake, 0, len(questions))
	for _, q := range questions {
		correct := ""
		for _, opt := range q.Options {
			if opt.IsCorrect {
				correct = opt.Content
				break
			}
		}
		result = append(result, StudentMistake{
			QuestionID:    q.ID,
			Title:         q.Title,
			WrongCount:    wrongCountMap[q.ID],
			CorrectOption: correct,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].WrongCount > result[j].WrongCount
	})
	return result, nil
}

func (s *AttemptService) ClassWrongStats() ([]ClassWrongStat, error) {
	var wrongAnswers []models.AttemptAnswer
	if err := s.db.Where("is_correct = ?", false).Find(&wrongAnswers).Error; err != nil {
		return nil, fmt.Errorf("load wrong answers: %w", err)
	}
	if len(wrongAnswers) == 0 {
		return []ClassWrongStat{}, nil
	}

	attemptIDSet := map[uint]struct{}{}
	questionIDSet := map[uint]struct{}{}
	for _, item := range wrongAnswers {
		attemptIDSet[item.AttemptID] = struct{}{}
		questionIDSet[item.QuestionID] = struct{}{}
	}

	attemptIDs := mapKeys(attemptIDSet)
	questionIDs := mapKeys(questionIDSet)

	var attempts []models.Attempt
	if err := s.db.Preload("ClassRoom").Where("id IN ?", attemptIDs).Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("load attempts for stats: %w", err)
	}
	attemptMap := map[uint]models.Attempt{}
	for _, a := range attempts {
		attemptMap[a.ID] = a
	}

	var questions []models.Question
	if err := s.db.Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load questions for stats: %w", err)
	}
	questionMap := map[uint]models.Question{}
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	type statKey struct {
		classID    uint
		questionID uint
	}
	counter := map[statKey]int64{}
	for _, item := range wrongAnswers {
		attempt, ok := attemptMap[item.AttemptID]
		if !ok {
			continue
		}
		key := statKey{classID: attempt.ClassID, questionID: item.QuestionID}
		counter[key]++
	}

	result := make([]ClassWrongStat, 0, len(counter))
	for key, count := range counter {
		attemptClass := ""
		if at, ok := findAttemptByClass(attempts, key.classID); ok {
			attemptClass = at.ClassRoom.Name
		}
		question := questionMap[key.questionID]
		result = append(result, ClassWrongStat{
			ClassID:    key.classID,
			ClassName:  attemptClass,
			QuestionID: key.questionID,
			Question:   question.Title,
			WrongCount: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].WrongCount > result[j].WrongCount
	})
	return result, nil
}

func (s *AttemptService) TeacherRecentAttempts(limit int) ([]RecentAttempt, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var attempts []models.Attempt
	if err := s.db.Preload("User").Preload("ClassRoom").Order("created_at desc").Limit(limit).Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("load recent attempts: %w", err)
	}

	result := make([]RecentAttempt, 0, len(attempts))
	for _, item := range attempts {
		result = append(result, RecentAttempt{
			ID:        item.ID,
			Student:   item.User.Username,
			ClassName: item.ClassRoom.Name,
			Score:     item.Score,
			Total:     item.Total,
			CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *AttemptService) Overview() (*Overview, error) {
	result := &Overview{}
	if err := s.db.Model(&models.User{}).Where("role = ?", models.RoleStudent).Count(&result.StudentCount).Error; err != nil {
		return nil, fmt.Errorf("count students: %w", err)
	}
	if err := s.db.Model(&models.ClassRoom{}).Count(&result.ClassCount).Error; err != nil {
		return nil, fmt.Errorf("count classes: %w", err)
	}
	if err := s.db.Model(&models.Question{}).Count(&result.QuestionCount).Error; err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}
	if err := s.db.Model(&models.Attempt{}).Count(&result.AttemptCount).Error; err != nil {
		return nil, fmt.Errorf("count attempts: %w", err)
	}
	return result, nil
}

func uniqueQuestionIDs(items []dto.SubmitAnswerItem) []uint {
	set := map[uint]struct{}{}
	for _, item := range items {
		set[item.QuestionID] = struct{}{}
	}
	ids := make([]uint, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	return ids
}

func mapKeys(set map[uint]struct{}) []uint {
	keys := make([]uint, 0, len(set))
	for id := range set {
		keys = append(keys, id)
	}
	return keys
}

func findAttemptByClass(attempts []models.Attempt, classID uint) (models.Attempt, bool) {
	for _, item := range attempts {
		if item.ClassID == classID {
			return item, true
		}
	}
	return models.Attempt{}, false
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
