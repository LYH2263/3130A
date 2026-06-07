package service

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"

	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

const MasteryStreakRequired = 2

type MistakeReviewService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewMistakeReviewService(db *gorm.DB, log *slog.Logger) *MistakeReviewService {
	return &MistakeReviewService{db: db, log: log}
}

func (s *MistakeReviewService) GetMistakeReviews(userID uint) ([]dto.MistakeReviewItem, error) {
	var attempts []models.Attempt
	if err := s.db.Where("user_id = ?", userID).Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("load attempts: %w", err)
	}
	if len(attempts) == 0 {
		return []dto.MistakeReviewItem{}, nil
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
		return []dto.MistakeReviewItem{}, nil
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

	var reviews []models.MistakeReview
	if err := s.db.Where("user_id = ? AND question_id IN ?", userID, questionIDs).Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("load mistake reviews: %w", err)
	}
	reviewMap := map[uint]models.MistakeReview{}
	for _, r := range reviews {
		reviewMap[r.QuestionID] = r
	}

	result := make([]dto.MistakeReviewItem, 0, len(questions))
	for _, q := range questions {
		review := reviewMap[q.ID]
		correct := getCorrectAnswer(q)
		masteryRate := 0
		if review.Status == models.MistakeReviewStatusMastered {
			masteryRate = 100
		} else {
			masteryRate = int(math.Min(float64(review.StreakCorrect)/float64(MasteryStreakRequired)*100, 100))
		}

		status := review.Status
		if status == "" {
			status = models.MistakeReviewStatusPending
		}

		result = append(result, dto.MistakeReviewItem{
			QuestionID:    q.ID,
			Title:         q.Title,
			WrongCount:    wrongCountMap[q.ID],
			CorrectOption: correct,
			Type:          q.Type,
			Status:        status,
			ReviewCount:   review.ReviewCount,
			StreakCorrect: review.StreakCorrect,
			MasteryRate:   masteryRate,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Status != result[j].Status {
			return result[i].Status == models.MistakeReviewStatusPending
		}
		return result[i].WrongCount > result[j].WrongCount
	})

	return result, nil
}

func (s *MistakeReviewService) GenerateReviewQuiz(userID uint, limit int) ([]models.Question, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	mistakes, err := s.GetMistakeReviews(userID)
	if err != nil {
		return nil, err
	}

	pendingIDs := make([]uint, 0)
	for _, m := range mistakes {
		if m.Status != models.MistakeReviewStatusMastered {
			pendingIDs = append(pendingIDs, m.QuestionID)
		}
	}

	if len(pendingIDs) == 0 {
		return []models.Question{}, nil
	}

	targetCount := limit
	if len(pendingIDs) < targetCount {
		targetCount = len(pendingIDs)
	}
	selectedIDs := pendingIDs[:targetCount]

	var questions []models.Question
	if err := s.db.Preload("Options").Preload("BlankAnswers").Where("id IN ?", selectedIDs).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("load review questions: %w", err)
	}

	questionMap := make(map[uint]models.Question, len(questions))
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	result := make([]models.Question, 0, len(selectedIDs))
	for _, id := range selectedIDs {
		if q, ok := questionMap[id]; ok {
			result = append(result, q)
		}
	}

	return result, nil
}

func (s *MistakeReviewService) SubmitReview(userID uint, classID uint, req dto.SubmitRequest) (*dto.MistakeReviewResult, error) {
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

	var reviews []models.MistakeReview
	if err := s.db.Where("user_id = ? AND question_id IN ?", userID, questionIDs).Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("load reviews: %w", err)
	}
	reviewMap := map[uint]*models.MistakeReview{}
	for i := range reviews {
		reviewMap[reviews[i].QuestionID] = &reviews[i]
	}

	details := make([]dto.MistakeReviewAnswerDetail, 0, len(req.Answers))
	totalScore := 0
	totalMaxScore := 0
	newlyMastered := make([]dto.MistakeReviewAnswerDetail, 0)
	stillNeedReview := make([]dto.MistakeReviewAnswerDetail, 0)

	now := time.Now()

	for _, answer := range req.Answers {
		question, ok := questionMap[answer.QuestionID]
		if !ok {
			return nil, ErrInvalidSubmission
		}

		score, maxScore, isCorrect := gradeQuestion(question, answer)
		totalScore += score
		totalMaxScore += maxScore

		review, exists := reviewMap[answer.QuestionID]
		if !exists {
			review = &models.MistakeReview{
				UserID:     userID,
				QuestionID: answer.QuestionID,
				Status:     models.MistakeReviewStatusPending,
			}
		}

		wasMastered := review.Status == models.MistakeReviewStatusMastered
		isNewlyMastered := false

		if !wasMastered {
			review.ReviewCount++
			review.LastReviewedAt = &now

			if isCorrect {
				review.StreakCorrect++
				if review.StreakCorrect >= MasteryStreakRequired {
					review.Status = models.MistakeReviewStatusMastered
					isNewlyMastered = true
				}
			} else {
				review.StreakCorrect = 0
			}
		}

		detail := dto.MistakeReviewAnswerDetail{
			QuestionID:    answer.QuestionID,
			Score:         score,
			MaxScore:      maxScore,
			IsCorrect:     isCorrect,
			Type:          question.Type,
			IsNewlyMastered: isNewlyMastered,
			WasMastered:   wasMastered,
			ReviewCount:   review.ReviewCount,
			Status:        review.Status,
		}
		details = append(details, detail)

		if isNewlyMastered {
			newlyMastered = append(newlyMastered, detail)
		} else if !wasMastered && !isCorrect {
			stillNeedReview = append(stillNeedReview, detail)
		} else if !wasMastered && isCorrect {
			stillNeedReview = append(stillNeedReview, detail)
		}
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin transaction: %w", tx.Error)
	}

	for _, answer := range req.Answers {
		review := reviewMap[answer.QuestionID]
		if review.ID == 0 {
			if err := tx.Create(review).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("create review: %w", err)
			}
		} else {
			if err := tx.Save(review).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("update review: %w", err)
			}
		}
	}

	answersModel := make([]models.AttemptAnswer, 0, len(req.Answers))
	for _, answer := range req.Answers {
		question := questionMap[answer.QuestionID]
		score, maxScore, isCorrect := gradeQuestion(question, answer)
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
	}

	attempt := models.Attempt{
		UserID:  userID,
		ClassID: classID,
		Score:   totalScore,
		Total:   totalMaxScore,
		Answers: answersModel,
	}
	if err := tx.Create(&attempt).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("save attempt: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	rate := fmt.Sprintf("%.0f%%", (float64(totalScore)/float64(totalMaxScore))*100)
	s.log.Info("mistake review submitted", "userID", userID, "score", totalScore, "total", totalMaxScore, "newlyMastered", len(newlyMastered))

	return &dto.MistakeReviewResult{
		Score:           totalScore,
		Total:           totalMaxScore,
		Rate:            rate,
		NewlyMastered:   newlyMastered,
		StillNeedReview: stillNeedReview,
		Details:         details,
	}, nil
}
