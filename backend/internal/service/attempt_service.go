package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

type AttemptService struct {
	db  *gorm.DB
	log *slog.Logger
}

type SubmitResult struct {
	AttemptID uint             `json:"attemptId"`
	Score     int              `json:"score"`
	Total     int              `json:"total"`
	Rate      string           `json:"rate"`
	Details   []dto.AnswerDetail `json:"details"`
}

type StudentMistake struct {
	QuestionID    uint   `json:"questionId"`
	Title         string `json:"title"`
	WrongCount    int64  `json:"wrongCount"`
	CorrectOption string `json:"correctOption"`
	Type          string `json:"type"`
}

type ClassWrongStat struct {
	ClassID    uint   `json:"classId"`
	ClassName  string `json:"className"`
	QuestionID uint   `json:"questionId"`
	Question   string `json:"question"`
	WrongCount int64  `json:"wrongCount"`
	Type       string `json:"type"`
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
	if err := s.db.Preload("Options").Preload("BlankAnswers").Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load questions: %w", err)
	}
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	questionMap := make(map[uint]models.Question, len(questions))
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	answerMap := make(map[uint]dto.SubmitAnswerItem, len(req.Answers))
	for _, ans := range req.Answers {
		answerMap[ans.QuestionID] = ans
	}

	answersModel := make([]models.AttemptAnswer, 0, len(req.Answers))
	details := make([]dto.AnswerDetail, 0, len(req.Answers))
	totalScore := 0
	totalMaxScore := 0

	for _, answer := range req.Answers {
		question, ok := questionMap[answer.QuestionID]
		if !ok {
			return nil, ErrInvalidSubmission
		}

		score, maxScore, isCorrect := gradeQuestion(question, answer)

		totalScore += score
		totalMaxScore += maxScore

		ansModel := models.AttemptAnswer{
			QuestionID:   answer.QuestionID,
			QuestionType: question.Type,
			IsCorrect:    isCorrect,
			Score:        score,
			MaxScore:     maxScore,
		}

		switch question.Type {
		case models.QuestionTypeSingle, models.QuestionTypeJudge:
			ansModel.SelectedOptionID = answer.OptionID
		case models.QuestionTypeMultiple:
			ansModel.SelectedOptionIDs = models.UintArray(answer.OptionIDs)
		case models.QuestionTypeBlank:
			ansModel.BlankAnswer = answer.BlankAnswer
		}

		answersModel = append(answersModel, ansModel)
		details = append(details, dto.AnswerDetail{
			QuestionID: answer.QuestionID,
			Score:      score,
			MaxScore:   maxScore,
			IsCorrect:  isCorrect,
			Type:       question.Type,
		})
	}

	attempt := models.Attempt{
		UserID:  userID,
		ClassID: classID,
		Score:   totalScore,
		Total:   totalMaxScore,
		Answers: answersModel,
	}
	if err := s.db.Create(&attempt).Error; err != nil {
		return nil, fmt.Errorf("save attempt: %w", err)
	}

	rate := fmt.Sprintf("%.0f%%", (float64(totalScore)/float64(totalMaxScore))*100)
	s.log.Info("attempt submitted", "attemptID", attempt.ID, "userID", userID, "score", totalScore, "total", totalMaxScore)

	return &SubmitResult{
		AttemptID: attempt.ID,
		Score:     totalScore,
		Total:     totalMaxScore,
		Rate:      rate,
		Details:   details,
	}, nil
}

func gradeQuestion(question models.Question, answer dto.SubmitAnswerItem) (int, int, bool) {
	maxScore := 100
	switch question.Type {
	case models.QuestionTypeSingle, models.QuestionTypeJudge:
		score := gradeSingleOrJudge(question, answer)
		return score, maxScore, score == maxScore
	case models.QuestionTypeMultiple:
		score := gradeMultiple(question, answer)
		return score, maxScore, score == maxScore
	case models.QuestionTypeBlank:
		score := gradeBlank(question, answer)
		return score, maxScore, score == maxScore
	default:
		return 0, maxScore, false
	}
}

func gradeSingleOrJudge(question models.Question, answer dto.SubmitAnswerItem) int {
	for _, opt := range question.Options {
		if opt.ID == answer.OptionID {
			if opt.IsCorrect {
				return 100
			}
			break
		}
	}
	return 0
}

func gradeMultiple(question models.Question, answer dto.SubmitAnswerItem) int {
	scoring := question.MultipleScore
	if scoring == "" {
		scoring = models.MultipleScoringAllOrNothing
	}

	correctIDs := make(map[uint]bool)
	correctCount := 0
	for _, opt := range question.Options {
		if opt.IsCorrect {
			correctIDs[opt.ID] = true
			correctCount++
		}
	}

	selectedSet := make(map[uint]bool)
	selectedCorrectCount := 0
	for _, id := range answer.OptionIDs {
		selectedSet[id] = true
		if correctIDs[id] {
			selectedCorrectCount++
		}
	}

	hasWrongSelection := false
	for _, id := range answer.OptionIDs {
		if !correctIDs[id] {
			hasWrongSelection = true
			break
		}
	}

	if scoring == models.MultipleScoringAllOrNothing {
		if selectedCorrectCount == correctCount && len(answer.OptionIDs) == correctCount && !hasWrongSelection {
			return 100
		}
		return 0
	}

	if hasWrongSelection {
		return 0
	}

	if correctCount == 0 {
		return 0
	}

	ratio := float64(selectedCorrectCount) / float64(correctCount)
	score := int(math.Round(ratio * 100))
	return score
}

func gradeBlank(question models.Question, answer dto.SubmitAnswerItem) int {
	userAnswer := normalizeBlankAnswer(answer.BlankAnswer)
	if userAnswer == "" {
		return 0
	}

	for _, ba := range question.BlankAnswers {
		correctAns := normalizeBlankAnswer(ba.Answer)
		matchMode := ba.MatchMode
		if matchMode == "" {
			matchMode = models.BlankMatchExact
		}

		switch matchMode {
		case models.BlankMatchExact:
			if userAnswer == correctAns {
				return 100
			}
		case models.BlankMatchIgnoreCase:
			if strings.EqualFold(userAnswer, correctAns) {
				return 100
			}
		case models.BlankMatchRegex:
			re, err := regexp.Compile(correctAns)
			if err != nil {
				continue
			}
			if re.MatchString(userAnswer) {
				return 100
			}
		}
	}

	return 0
}

func normalizeBlankAnswer(s string) string {
	s = strings.TrimSpace(s)
	s = fullWidthToHalfWidth(s)
	return s
}

func fullWidthToHalfWidth(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r >= 0xFF01 && r <= 0xFF5E {
			builder.WriteRune(r - 0xFEE0)
		} else if r == 0x3000 {
			builder.WriteRune(0x20)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
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
	if err := s.db.Preload("Options").Preload("BlankAnswers").Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load mistake questions: %w", err)
	}

	result := make([]StudentMistake, 0, len(questions))
	for _, q := range questions {
		correct := getCorrectAnswer(q)
		result = append(result, StudentMistake{
			QuestionID:    q.ID,
			Title:         q.Title,
			WrongCount:    wrongCountMap[q.ID],
			CorrectOption: correct,
			Type:          q.Type,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].WrongCount > result[j].WrongCount
	})
	return result, nil
}

func getCorrectAnswer(q models.Question) string {
	switch q.Type {
	case models.QuestionTypeSingle, models.QuestionTypeJudge, models.QuestionTypeMultiple:
		var corrects []string
		for _, opt := range q.Options {
			if opt.IsCorrect {
				corrects = append(corrects, opt.Content)
			}
		}
		return strings.Join(corrects, "、")
	case models.QuestionTypeBlank:
		var corrects []string
		for _, ba := range q.BlankAnswers {
			corrects = append(corrects, ba.Answer)
		}
		return strings.Join(corrects, " / ")
	default:
		return ""
	}
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
			Type:       question.Type,
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

func (s *AttemptService) SaveDraft(userID uint, req dto.SaveDraftRequest) error {
	questionsJSON, err := json.Marshal(req.Questions)
	if err != nil {
		return fmt.Errorf("marshal questions: %w", err)
	}
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return fmt.Errorf("marshal answers: %w", err)
	}

	var draft models.AttemptDraft
	err = s.db.Where("user_id = ? AND quiz_mode = ?", userID, req.QuizMode).First(&draft).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find draft: %w", err)
	}

	if draft.ID == 0 {
		draft = models.AttemptDraft{
			UserID:       userID,
			QuizMode:     req.QuizMode,
			QuestionData: string(questionsJSON),
			AnswerData:   string(answersJSON),
		}
		if err := s.db.Create(&draft).Error; err != nil {
			return fmt.Errorf("create draft: %w", err)
		}
	} else {
		draft.QuestionData = string(questionsJSON)
		draft.AnswerData = string(answersJSON)
		if err := s.db.Save(&draft).Error; err != nil {
			return fmt.Errorf("update draft: %w", err)
		}
	}

	s.log.Info("draft saved", "userID", userID, "quizMode", req.QuizMode)
	return nil
}

func (s *AttemptService) GetDraft(userID uint, quizMode string) (*dto.DraftResponse, error) {
	var draft models.AttemptDraft
	err := s.db.Where("user_id = ? AND quiz_mode = ?", userID, quizMode).First(&draft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDraftNotFound
		}
		return nil, fmt.Errorf("find draft: %w", err)
	}

	var questions []map[string]interface{}
	if err := json.Unmarshal([]byte(draft.QuestionData), &questions); err != nil {
		return nil, fmt.Errorf("unmarshal questions: %w", err)
	}

	var answers map[string]interface{}
	if err := json.Unmarshal([]byte(draft.AnswerData), &answers); err != nil {
		return nil, fmt.Errorf("unmarshal answers: %w", err)
	}

	return &dto.DraftResponse{
		QuizMode:  draft.QuizMode,
		Questions: questions,
		Answers:   answers,
		UpdatedAt: draft.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *AttemptService) ClearDraft(userID uint, quizMode string) error {
	result := s.db.Where("user_id = ? AND quiz_mode = ?", userID, quizMode).Delete(&models.AttemptDraft{})
	if result.Error != nil {
		return fmt.Errorf("delete draft: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		s.log.Info("draft cleared", "userID", userID, "quizMode", quizMode)
	}
	return nil
}

func (s *AttemptService) GetAttemptDetail(userID uint, attemptID uint) (*dto.AttemptReport, error) {
	var attempt models.Attempt
	if err := s.db.Preload("Answers").Where("id = ?", attemptID).First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttemptNotFound
		}
		return nil, fmt.Errorf("load attempt: %w", err)
	}

	if attempt.UserID != userID {
		return nil, ErrAttemptForbidden
	}

	answerIDs := make([]uint, 0, len(attempt.Answers))
	questionIDSet := map[uint]struct{}{}
	for _, a := range attempt.Answers {
		answerIDs = append(answerIDs, a.ID)
		questionIDSet[a.QuestionID] = struct{}{}
	}

	questionIDs := make([]uint, 0, len(questionIDSet))
	for id := range questionIDSet {
		questionIDs = append(questionIDs, id)
	}

	var questions []models.Question
	if err := s.db.Preload("Options").Preload("BlankAnswers").Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load questions: %w", err)
	}
	questionMap := map[uint]models.Question{}
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	answers := make([]dto.AttemptReportAnswer, 0, len(attempt.Answers))
	correctCount := 0
	wrongCount := 0

	for _, ans := range attempt.Answers {
		q, ok := questionMap[ans.QuestionID]
		if !ok {
			continue
		}

		reportAns := dto.AttemptReportAnswer{
			QuestionID:    ans.QuestionID,
			QuestionTitle: q.Title,
			QuestionType:  ans.QuestionType,
			Score:         ans.Score,
			MaxScore:      ans.MaxScore,
			IsCorrect:     ans.IsCorrect,
		}

		if ans.IsCorrect {
			correctCount++
		} else {
			wrongCount++
		}

		switch ans.QuestionType {
		case models.QuestionTypeSingle, models.QuestionTypeJudge, models.QuestionTypeMultiple:
			opts := make([]dto.AttemptReportOption, 0, len(q.Options))
			for _, opt := range q.Options {
				opts = append(opts, dto.AttemptReportOption{
					ID:        opt.ID,
					Content:   opt.Content,
					IsCorrect: opt.IsCorrect,
				})
			}
			reportAns.Options = opts

			if ans.QuestionType == models.QuestionTypeSingle || ans.QuestionType == models.QuestionTypeJudge {
				reportAns.SelectedOptionID = ans.SelectedOptionID
			} else {
				reportAns.SelectedOptionIDs = []uint(ans.SelectedOptionIDs)
			}
		case models.QuestionTypeBlank:
			reportAns.BlankAnswer = ans.BlankAnswer
			correctAnswers := make([]string, 0, len(q.BlankAnswers))
			for _, ba := range q.BlankAnswers {
				correctAnswers = append(correctAnswers, ba.Answer)
			}
			reportAns.CorrectBlankAnswers = correctAnswers
		}

		answers = append(answers, reportAns)
	}

	rate := "0%"
	if attempt.Total > 0 {
		rate = fmt.Sprintf("%.0f%%", (float64(attempt.Score)/float64(attempt.Total))*100)
	}

	return &dto.AttemptReport{
		ID:            attempt.ID,
		Score:         attempt.Score,
		Total:         attempt.Total,
		Rate:          rate,
		CorrectCount:  correctCount,
		WrongCount:    wrongCount,
		QuestionCount: len(answers),
		CreatedAt:     attempt.CreatedAt.Format("2006-01-02 15:04:05"),
		Answers:       answers,
	}, nil
}
