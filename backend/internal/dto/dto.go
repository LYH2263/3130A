package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	ClassID  uint   `json:"classId" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type QuestionOptionInput struct {
	Content   string `json:"content" binding:"min=0,max=200"`
	IsCorrect bool   `json:"isCorrect"`
}

type BlankAnswerInput struct {
	Answer    string `json:"answer" binding:"min=0,max=500"`
	MatchMode string `json:"matchMode" binding:"omitempty,oneof=exact ignore_case regex"`
}

type QuestionInput struct {
	Type                 string                `json:"type" binding:"required,oneof=single multiple judge blank"`
	Title                string                `json:"title" binding:"required,min=2,max=1000"`
	Description          string                `json:"description" binding:"max=2000"`
	CategoryID           *uint                 `json:"categoryId"`
	TagNames             []string              `json:"tagNames"`
	KnowledgePointNames  []string              `json:"knowledgePointNames"`
	ExplanationContent   string                `json:"explanationContent" binding:"max=5000"`
	ExplanationRefs      string                `json:"explanationRefs" binding:"max=2000"`
	Options              []QuestionOptionInput `json:"options" binding:"dive"`
	BlankAnswers         []BlankAnswerInput    `json:"blankAnswers" binding:"dive"`
	MultipleScore        string                `json:"multipleScore" binding:"omitempty,oneof=all_or_nothing partial"`
}

type QuestionQuery struct {
	Keyword            string   `form:"keyword"`
	CategoryID         *uint    `form:"categoryId"`
	TagIDs             []uint   `form:"tagIds"`
	TagMode            string   `form:"tagMode" binding:"omitempty,oneof=and or"`
	CreatedBy          *uint    `form:"createdBy"`
	CreatedFrom        string   `form:"createdFrom"`
	CreatedTo          string   `form:"createdTo"`
	HasAnswerError     *bool    `form:"hasAnswerError"`
	SortBy             string   `form:"sortBy" binding:"omitempty,oneof=created_at id wrong_count"`
	SortOrder          string   `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page               int      `form:"page,default=1"`
	PageSize           int      `form:"pageSize,default=20"`
}

type PaginatedQuestions struct {
	Items    []QuestionDetail `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type QuestionDetail struct {
	ID                   uint     `json:"id"`
	Type                 string   `json:"type"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	CategoryID           *uint    `json:"categoryId"`
	CategoryName         string   `json:"categoryName"`
	CreatedBy            uint     `json:"createdBy"`
	CreatedByName        string   `json:"createdByName"`
	MultipleScore        string   `json:"multipleScore"`
	ExplanationContent   string   `json:"explanationContent"`
	ExplanationRefs      string   `json:"explanationRefs"`
	KnowledgePoints      []KnowledgePointInfo `json:"knowledgePoints"`
	WrongCount           int64    `json:"wrongCount"`
	HasAnswerError       bool     `json:"hasAnswerError"`
	CreatedAt            string   `json:"createdAt"`
	UpdatedAt            string   `json:"updatedAt"`
}

type KnowledgePointInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type KnowledgePointInput struct {
	Name string `json:"name" binding:"required,max=128"`
	Sort int    `json:"sort"`
}

type QuestionExplanationDTO struct {
	Content         string               `json:"content"`
	References    string               `json:"references"`
	KnowledgePoints []KnowledgePointInfo `json:"knowledgePoints"`
}

type CategoryInput struct {
	Name     string `json:"name" binding:"required,max=128"`
	ParentID *uint  `json:"parentId"`
	Sort     int    `json:"sort"`
}

type TagInput struct {
	Name string `json:"name" binding:"required,max=64"`
}

type UploadQuestionPayload struct {
	Questions []QuestionInput `json:"questions"`
}

type SubmitAnswerItem struct {
	QuestionID       uint     `json:"questionId" binding:"required"`
	OptionID         uint     `json:"optionId"`
	OptionIDs        []uint   `json:"optionIds"`
	BlankAnswer      string   `json:"blankAnswer"`
}

type SubmitRequest struct {
	Answers []SubmitAnswerItem `json:"answers" binding:"required,min=1,dive"`
}

type AnswerDetail struct {
	QuestionID uint   `json:"questionId"`
	Score      int    `json:"score"`
	MaxScore   int    `json:"maxScore"`
	IsCorrect  bool   `json:"isCorrect"`
	Status     string `json:"status"`
	Type       string `json:"type"`
}

const (
	AnswerStatusCorrect = "correct"
	AnswerStatusPartial = "partial"
	AnswerStatusWrong   = "wrong"
)

type SubmitResultDetail struct {
	AttemptID uint           `json:"attemptId"`
	Score     int            `json:"score"`
	Total     int            `json:"total"`
	Rate      string         `json:"rate"`
	Details   []AnswerDetail `json:"details"`
}

type MistakeReviewItem struct {
	QuestionID       uint               `json:"questionId"`
	Title            string             `json:"title"`
	WrongCount       int64              `json:"wrongCount"`
	CorrectOption    string             `json:"correctOption"`
	Type             string             `json:"type"`
	Status           string             `json:"status"`
	ReviewCount      int                `json:"reviewCount"`
	StreakCorrect    int                `json:"streakCorrect"`
	MasteryRate      int                `json:"masteryRate"`
	ExplanationContent  string          `json:"explanationContent"`
	ExplanationRefs    string          `json:"explanationRefs"`
	KnowledgePoints  []KnowledgePointInfo `json:"knowledgePoints"`
}

type MistakeReviewSubmitRequest struct {
	Answers []SubmitAnswerItem `json:"answers" binding:"required,min=1,dive"`
}

type MistakeReviewAnswerDetail struct {
	QuestionID    uint   `json:"questionId"`
	Score         int    `json:"score"`
	MaxScore      int    `json:"maxScore"`
	IsCorrect     bool   `json:"isCorrect"`
	Type          string `json:"type"`
	IsNewlyMastered bool `json:"isNewlyMastered"`
	WasMastered   bool   `json:"wasMastered"`
	ReviewCount   int    `json:"reviewCount"`
	Status        string `json:"status"`
}

type MistakeReviewResult struct {
	Score            int                       `json:"score"`
	Total            int                       `json:"total"`
	Rate             string                    `json:"rate"`
	NewlyMastered    []MistakeReviewAnswerDetail `json:"newlyMastered"`
	StillNeedReview  []MistakeReviewAnswerDetail `json:"stillNeedReview"`
	Details          []MistakeReviewAnswerDetail `json:"details"`
	SkippedCount     int                       `json:"skippedCount"`
}

type SaveDraftRequest struct {
	QuizMode    string                   `json:"quizMode" binding:"required,oneof=normal review"`
	Questions   []map[string]interface{} `json:"questions" binding:"required,min=1"`
	Answers     map[string]interface{}   `json:"answers" binding:"required"`
	LastUpdated string                   `json:"lastUpdated"`
	Force       bool                     `json:"force"`
}

type DraftResponse struct {
	QuizMode  string                   `json:"quizMode"`
	Questions []map[string]interface{} `json:"questions"`
	Answers   map[string]interface{}   `json:"answers"`
	UpdatedAt string                   `json:"updatedAt"`
}

type FavoriteItem struct {
	ID          uint   `json:"id"`
	QuestionID  uint   `json:"questionId"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	CreatedAt   string `json:"createdAt"`
}

type QuestionSetDTO struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	QuestionIDs   []uint `json:"questionIds"`
	QuestionCount int    `json:"questionCount"`
	SortOrder     int    `json:"sortOrder"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type CreateQuestionSetRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"max=500"`
	QuestionIDs []uint `json:"questionIds"`
}

type UpdateQuestionSetRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=128"`
	Description string `json:"description" binding:"omitempty,max=500"`
	QuestionIDs []uint `json:"questionIds"`
	SortOrder   *int   `json:"sortOrder"`
}

type AddQuestionsToSetRequest struct {
	QuestionIDs []uint `json:"questionIds" binding:"required,min=1"`
}

type ReorderSetQuestionsRequest struct {
	QuestionIDs []uint `json:"questionIds" binding:"required"`
}

type BatchFavoriteStatusRequest struct {
	QuestionIDs []uint `json:"questionIds"`
}

type AttemptReportOption struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	IsCorrect bool   `json:"isCorrect"`
}

type AttemptReportAnswer struct {
	QuestionID       uint                 `json:"questionId"`
	QuestionTitle    string               `json:"questionTitle"`
	QuestionType     string               `json:"questionType"`
	Score            int                  `json:"score"`
	MaxScore         int                  `json:"maxScore"`
	IsCorrect        bool                 `json:"isCorrect"`
	Options          []AttemptReportOption `json:"options,omitempty"`
	SelectedOptionID uint                 `json:"selectedOptionId,omitempty"`
	SelectedOptionIDs []uint              `json:"selectedOptionIds,omitempty"`
	BlankAnswer      string               `json:"blankAnswer,omitempty"`
	CorrectBlankAnswers []string          `json:"correctBlankAnswers,omitempty"`
}

type AttemptReport struct {
	ID             uint                  `json:"id"`
	Score          int                   `json:"score"`
	Total          int                   `json:"total"`
	Rate           string                `json:"rate"`
	CorrectCount   int                   `json:"correctCount"`
	WrongCount     int                   `json:"wrongCount"`
	QuestionCount  int                   `json:"questionCount"`
	CreatedAt      string                `json:"createdAt"`
	Answers        []AttemptReportAnswer `json:"answers"`
}

type LeaderboardQuery struct {
	ScoreType   string `form:"scoreType,default=highest" binding:"omitempty,oneof=highest average weighted"`
	ClassID     *uint  `form:"classId"`
	WeightedN   int    `form:"weightedN,default=5"`
	Limit       int    `form:"limit,default=50"`
	Page        int    `form:"page,default=1"`
}

type LeaderboardItem struct {
	Rank          int     `json:"rank"`
	UserID        uint    `json:"userId"`
	Username      string  `json:"username"`
	ClassID       uint    `json:"classId"`
	ClassName     string  `json:"className"`
	Score         float64 `json:"score"`
	ScoreDisplay  string  `json:"scoreDisplay"`
	AttemptCount  int     `json:"attemptCount"`
	CorrectRate   string  `json:"correctRate"`
	IsCurrentUser bool    `json:"isCurrentUser"`
}

type LeaderboardResult struct {
	Items       []LeaderboardItem `json:"items"`
	Total       int64             `json:"total"`
	ScoreType   string            `json:"scoreType"`
	ClassID     *uint             `json:"classId"`
	ClassName   string            `json:"className"`
	CurrentRank *LeaderboardItem  `json:"currentRank,omitempty"`
	GapToPrev   float64           `json:"gapToPrev"`
	HasPrev     bool              `json:"hasPrev"`
}
