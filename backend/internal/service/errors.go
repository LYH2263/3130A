package service

import "errors"

var (
	ErrUserExists        = errors.New("username already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrClassNotFound     = errors.New("class not found")
	ErrInvalidQuestion   = errors.New("question must contain exactly one correct option")
	ErrQuestionNotFound  = errors.New("question not found")
	ErrNoQuestions       = errors.New("question bank is empty")
	ErrInvalidSubmission = errors.New("invalid submission")
)
