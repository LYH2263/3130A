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
	Type          string                `json:"type" binding:"required,oneof=single multiple judge blank"`
	Title         string                `json:"title" binding:"required,min=2,max=1000"`
	Description   string                `json:"description" binding:"max=2000"`
	CategoryID    *uint                 `json:"categoryId"`
	TagNames      []string              `json:"tagNames"`
	Options       []QuestionOptionInput `json:"options" binding:"dive"`
	BlankAnswers  []BlankAnswerInput    `json:"blankAnswers" binding:"dive"`
	MultipleScore string                `json:"multipleScore" binding:"omitempty,oneof=all_or_nothing partial"`
}

type QuestionQuery struct {
	Keyword    string   `form:"keyword"`
	CategoryID *uint    `form:"categoryId"`
	TagIDs     []uint   `form:"tagIds"`
	TagMode    string   `form:"tagMode" binding:"omitempty,oneof=and or"`
	Page       int      `form:"page,default=1"`
	PageSize   int      `form:"pageSize,default=20"`
}

type PaginatedQuestions struct {
	Items    []QuestionDetail `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type QuestionDetail struct {
	ID            uint   `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	CategoryID    *uint  `json:"categoryId"`
	CategoryName  string `json:"categoryName"`
	CreatedBy     uint   `json:"createdBy"`
	MultipleScore string `json:"multipleScore"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
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
	Type       string `json:"type"`
}

type SubmitResultDetail struct {
	AttemptID uint           `json:"attemptId"`
	Score     int            `json:"score"`
	Total     int            `json:"total"`
	Rate      string         `json:"rate"`
	Details   []AnswerDetail `json:"details"`
}

type MistakeReviewItem struct {
	QuestionID    uint   `json:"questionId"`
	Title         string `json:"title"`
	WrongCount    int64  `json:"wrongCount"`
	CorrectOption string `json:"correctOption"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	ReviewCount   int    `json:"reviewCount"`
	StreakCorrect int    `json:"streakCorrect"`
	MasteryRate   int    `json:"masteryRate"`
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
}

type SaveDraftRequest struct {
	QuizMode  string                   `json:"quizMode" binding:"required,oneof=normal review"`
	Questions []map[string]interface{} `json:"questions" binding:"required,min=1"`
	Answers   map[string]interface{}   `json:"answers" binding:"required"`
}

type DraftResponse struct {
	QuizMode  string                   `json:"quizMode"`
	Questions []map[string]interface{} `json:"questions"`
	Answers   map[string]interface{}   `json:"answers"`
	UpdatedAt string                   `json:"updatedAt"`
}
