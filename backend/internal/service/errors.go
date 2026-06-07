package service

import "errors"

var (
	ErrUserExists        = errors.New("username already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrClassNotFound     = errors.New("class not found")
	ErrInvalidQuestion   = errors.New("invalid question")
	ErrQuestionNotFound  = errors.New("question not found")
	ErrNoQuestions       = errors.New("question bank is empty")
	ErrInvalidSubmission = errors.New("invalid submission")
	ErrInvalidQuestionType  = errors.New("invalid question type")
	ErrInvalidSingleOptionCount = errors.New("single choice question must have 2-6 options")
	ErrInvalidSingleCorrect = errors.New("single choice question must have exactly one correct answer")
	ErrInvalidMultipleOptionCount = errors.New("multiple choice question must have 2-6 options")
	ErrInvalidMultipleCorrect = errors.New("multiple choice question must have at least two correct answers")
	ErrInvalidJudgeOptionCount = errors.New("judge question must have exactly two options")
	ErrInvalidJudgeCorrect = errors.New("judge question must have exactly one correct answer")
	ErrInvalidBlankAnswerCount = errors.New("blank question must have at least one correct answer")
	ErrInvalidBlankAnswer = errors.New("blank answer cannot be empty")
	ErrInvalidOptionContent = errors.New("option content cannot be empty")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrCategoryNameEmpty        = errors.New("category name cannot be empty")
	ErrCategoryParentInvalid = errors.New("invalid parent category")
	ErrTagNotFound          = errors.New("tag not found")
	ErrTagNameEmpty         = errors.New("tag name cannot be empty")
	ErrTagExists            = errors.New("tag already exists")
)
