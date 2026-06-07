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
	Options       []QuestionOptionInput `json:"options" binding:"dive"`
	BlankAnswers  []BlankAnswerInput    `json:"blankAnswers" binding:"dive"`
	MultipleScore string                `json:"multipleScore" binding:"omitempty,oneof=all_or_nothing partial"`
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
