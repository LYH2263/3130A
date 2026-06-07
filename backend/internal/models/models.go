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

	MistakeReviewStatusPending  = "pending"
	MistakeReviewStatusMastered = "mastered"

	LeaderboardScoreTypeHighest   = "highest"
	LeaderboardScoreTypeAverage   = "average"
	LeaderboardScoreTypeWeighted  = "weighted"

	DefaultLeaderboardLimit = 50
	DefaultWeightedRecentN = 5
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

type Category struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	ParentID  *uint      `gorm:"index" json:"parentId"`
	Name      string     `gorm:"size:128;not null" json:"name"`
	Sort      int        `gorm:"not null;default:0" json:"sort"`
	Children  []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type QuestionTag struct {
	QuestionID uint `gorm:"primaryKey;index" json:"questionId"`
	TagID      uint `gorm:"primaryKey;index" json:"tagId"`
}

type KnowledgePoint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;uniqueIndex;not null" json:"name"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type QuestionKnowledgePoint struct {
	QuestionID      uint `gorm:"primaryKey;index" json:"questionId"`
	KnowledgePointID uint `gorm:"primaryKey;index" json:"knowledgePointId"`
}

type QuestionExplanation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	QuestionID uint      `gorm:"uniqueIndex;not null" json:"questionId"`
	Content    string    `gorm:"type:text" json:"content"`
	References string    `gorm:"type:text" json:"references"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Question struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	Type              string           `gorm:"size:16;not null;default:'single';index" json:"type"`
	Title             string           `gorm:"type:text;not null" json:"title"`
	Description       string           `gorm:"type:text" json:"description"`
	CategoryID        *uint            `gorm:"index" json:"categoryId"`
	Category          *Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags              []Tag            `gorm:"many2many:question_tags;" json:"tags,omitempty"`
	KnowledgePoints   []KnowledgePoint `gorm:"many2many:question_knowledge_points;" json:"knowledgePoints,omitempty"`
	Explanation       *QuestionExplanation `gorm:"foreignKey:QuestionID" json:"explanation,omitempty"`
	CreatedBy         uint             `gorm:"index" json:"createdBy"`
	Options           []QuestionOption `json:"options,omitempty"`
	BlankAnswers      []BlankAnswer    `json:"blankAnswers,omitempty"`
	MultipleScore     string           `gorm:"size:20;not null;default:'all_or_nothing'" json:"multipleScore"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
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

type MistakeReview struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null;uniqueIndex:idx_user_question" json:"userId"`
	QuestionID      uint      `gorm:"index;not null;uniqueIndex:idx_user_question" json:"questionId"`
	Question        *Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Status          string    `gorm:"size:16;not null;default:'pending';index" json:"status"`
	ReviewCount     int       `gorm:"not null;default:0" json:"reviewCount"`
	StreakCorrect   int       `gorm:"not null;default:0" json:"streakCorrect"`
	LastReviewedAt  *time.Time `json:"lastReviewedAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AttemptDraft struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null;uniqueIndex:idx_user_mode" json:"userId"`
	QuizMode     string    `gorm:"size:16;not null;uniqueIndex:idx_user_mode" json:"quizMode"`
	QuestionData string    `gorm:"type:json;not null" json:"-"`
	AnswerData   string    `gorm:"type:json;not null" json:"-"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

const (
	QuizModeNormal = "normal"
	QuizModeReview = "review"
	QuizModeFavorite = "favorite"
	QuizModeSet = "set"
)

type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null;uniqueIndex:idx_user_question" json:"userId"`
	QuestionID uint      `gorm:"index;not null;uniqueIndex:idx_user_question" json:"questionId"`
	Question   *Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type QuestionSet struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"userId"`
	Name          string    `gorm:"size:128;not null" json:"name"`
	Description   string    `gorm:"size:500" json:"description"`
	QuestionIDs   UintArray `gorm:"type:json" json:"questionIds"`
	SortOrder     int       `gorm:"not null;default:0" json:"sortOrder"`
	QuestionCount int       `gorm:"-" json:"questionCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
