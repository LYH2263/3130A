package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	RoleTeacher = "teacher"
	RoleStudent = "student"

	QuestionTypeSingle   = "single"
	QuestionTypeMultiple = "multiple"
	QuestionTypeJudge    = "judge"
	QuestionTypeBlank    = "blank"

	BlankMatchExact       = "exact"
	BlankMatchIgnoreCase  = "ignore_case"
	BlankMatchRegex       = "regex"

	MultipleScoringAllOrNothing = "all_or_nothing"
	MultipleScoringPartial      = "partial"
)

type UintArray []uint

func (a UintArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *UintArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid data type for UintArray")
	}
	return json.Unmarshal(bytes, a)
}

type ClassRoom struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         string     `gorm:"size:16;not null;index" json:"role"`
	ClassID      *uint      `gorm:"index" json:"classId"`
	ClassRoom    *ClassRoom `gorm:"foreignKey:ClassID" json:"classRoom,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type Question struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	Type          string           `gorm:"size:16;not null;default:'single';index" json:"type"`
	Title         string           `gorm:"type:text;not null" json:"title"`
	Description   string           `gorm:"type:text" json:"description"`
	CreatedBy     uint             `gorm:"index" json:"createdBy"`
	Options       []QuestionOption `json:"options,omitempty"`
	BlankAnswers  []BlankAnswer    `json:"blankAnswers,omitempty"`
	MultipleScore string           `gorm:"size:20;not null;default:'all_or_nothing'" json:"multipleScore"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
}

type QuestionOption struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	QuestionID uint      `gorm:"index;not null" json:"questionId"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	IsCorrect  bool      `gorm:"not null" json:"isCorrect"`
	SortOrder  int       `gorm:"not null;default:0" json:"sortOrder"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type BlankAnswer struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	QuestionID  uint      `gorm:"index;not null" json:"questionId"`
	Answer      string    `gorm:"type:text;not null" json:"answer"`
	MatchMode   string    `gorm:"size:20;not null;default:'exact'" json:"matchMode"`
	SortOrder   int       `gorm:"not null;default:0" json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Attempt struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"index;not null" json:"userId"`
	User      User            `gorm:"foreignKey:UserID" json:"user"`
	ClassID   uint            `gorm:"index;not null" json:"classId"`
	ClassRoom ClassRoom       `gorm:"foreignKey:ClassID" json:"classRoom"`
	Score     int             `gorm:"not null" json:"score"`
	Total     int             `gorm:"not null" json:"total"`
	Answers   []AttemptAnswer `gorm:"foreignKey:AttemptID" json:"answers"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type AttemptAnswer struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	AttemptID          uint      `gorm:"index;not null" json:"attemptId"`
	QuestionID         uint      `gorm:"index;not null" json:"questionId"`
	QuestionType       string    `gorm:"size:16;not null" json:"questionType"`
	SelectedOptionID   uint      `gorm:"index" json:"selectedOptionId,omitempty"`
	SelectedOptionIDs  UintArray `gorm:"type:json" json:"selectedOptionIds,omitempty"`
	BlankAnswer        string    `gorm:"type:text" json:"blankAnswer,omitempty"`
	IsCorrect          bool      `gorm:"index;not null" json:"isCorrect"`
	Score              int       `gorm:"not null;default:0" json:"score"`
	MaxScore           int       `gorm:"not null;default:100" json:"maxScore"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
