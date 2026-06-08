package seed

import (
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"label3130/backend/internal/models"
)

func Run(db *gorm.DB, log *slog.Logger) error {
	classes := []string{"一班", "二班", "三班", "四班"}
	for _, name := range classes {
		if err := db.FirstOrCreate(&models.ClassRoom{}, models.ClassRoom{Name: name}).Error; err != nil {
			return fmt.Errorf("seed class %s: %w", name, err)
		}
	}

	if err := seedTeacher(db); err != nil {
		return err
	}
	if err := seedStudents(db); err != nil {
		return err
	}
	if err := seedQuestions(db); err != nil {
		return err
	}
	if err := seedExamConfigs(db); err != nil {
		return err
	}

	log.Info("seed completed")
	return nil
}

func seedTeacher(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.User{}).Where("role = ?", models.RoleTeacher).Count(&count).Error; err != nil {
		return fmt.Errorf("count teachers: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash teacher password: %w", err)
	}
	teacher := models.User{
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         models.RoleTeacher,
	}
	if err := db.Create(&teacher).Error; err != nil {
		return fmt.Errorf("create teacher: %w", err)
	}
	return nil
}

func seedStudents(db *gorm.DB) error {
	var classRoom models.ClassRoom
	if err := db.Where("name = ?", "一班").First(&classRoom).Error; err != nil {
		return fmt.Errorf("load class for student seed: %w", err)
	}

	students := []string{"stu001", "stu002"}
	for _, username := range students {
		var existing models.User
		err := db.Where("username = ?", username).First(&existing).Error
		if err == nil {
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("check student %s: %w", username, err)
		}

		hash, hashErr := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("hash student password: %w", hashErr)
		}
		user := models.User{
			Username:     username,
			PasswordHash: string(hash),
			Role:         models.RoleStudent,
			ClassID:      &classRoom.ID,
		}
		if createErr := db.Create(&user).Error; createErr != nil {
			return fmt.Errorf("create student %s: %w", username, createErr)
		}
	}
	return nil
}

func seedQuestions(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Question{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count questions: %w", err)
	}
	if count > 0 {
		return nil
	}

	templates := []models.Question{
		{
			Type:        models.QuestionTypeSingle,
			Title:       "TCP 三次握手中用于建立连接的第二步是？",
			Description: "网络基础",
			CreatedBy:   1,
			Options: []models.QuestionOption{
				{Content: "客户端发送 SYN", IsCorrect: false},
				{Content: "服务端返回 SYN+ACK", IsCorrect: true},
				{Content: "客户端发送 FIN", IsCorrect: false},
				{Content: "服务端直接发送 ACK", IsCorrect: false},
			},
		},
		{
			Type:        models.QuestionTypeSingle,
			Title:       "在 SQL 中用于去重查询结果的关键字是？",
			Description: "数据库基础",
			CreatedBy:   1,
			Options: []models.QuestionOption{
				{Content: "ORDER BY", IsCorrect: false},
				{Content: "UNIQUE", IsCorrect: false},
				{Content: "DISTINCT", IsCorrect: true},
				{Content: "GROUP", IsCorrect: false},
			},
		},
		{
			Type:        models.QuestionTypeJudge,
			Title:       "HTTP 状态码 404 表示资源未找到。",
			Description: "Web 基础",
			CreatedBy:   1,
			Options: []models.QuestionOption{
				{Content: "正确", IsCorrect: true},
				{Content: "错误", IsCorrect: false},
			},
		},
		{
			Type:        models.QuestionTypeMultiple,
			Title:       "以下哪些是 JavaScript 的基本数据类型？",
			Description: "JavaScript 基础",
			CreatedBy:   1,
			MultipleScore: models.MultipleScoringPartial,
			Options: []models.QuestionOption{
				{Content: "string", IsCorrect: true},
				{Content: "number", IsCorrect: true},
				{Content: "array", IsCorrect: false},
				{Content: "boolean", IsCorrect: true},
			},
		},
		{
			Type:        models.QuestionTypeBlank,
			Title:       "CSS 中用于设置元素背景颜色的属性是______。",
			Description: "CSS 基础",
			CreatedBy:   1,
			BlankAnswers: []models.BlankAnswer{
				{Answer: "background-color", MatchMode: models.BlankMatchIgnoreCase},
				{Answer: "background", MatchMode: models.BlankMatchExact},
			},
		},
		{
			Type:        models.QuestionTypeMultiple,
			Title:       "以下哪些是 Git 的常用命令？",
			Description: "开发工具",
			CreatedBy:   1,
			MultipleScore: models.MultipleScoringAllOrNothing,
			Options: []models.QuestionOption{
				{Content: "commit", IsCorrect: true},
				{Content: "push", IsCorrect: true},
				{Content: "compile", IsCorrect: false},
				{Content: "merge", IsCorrect: true},
			},
		},
		{
			Type:        models.QuestionTypeJudge,
			Title:       "HTML 是一种编程语言。",
			Description: "Web 基础",
			CreatedBy:   1,
			Options: []models.QuestionOption{
				{Content: "正确", IsCorrect: false},
				{Content: "错误", IsCorrect: true},
			},
		},
		{
			Type:        models.QuestionTypeBlank,
			Title:       "在 Go 语言中，声明变量使用的关键字是______。",
			Description: "Go 语言基础",
			CreatedBy:   1,
			BlankAnswers: []models.BlankAnswer{
				{Answer: "var", MatchMode: models.BlankMatchExact},
			},
		},
	}

	for _, item := range templates {
		if err := db.Create(&item).Error; err != nil {
			return fmt.Errorf("seed question: %w", err)
		}
	}
	return nil
}

func seedExamConfigs(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.ExamConfig{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count exam configs: %w", err)
	}
	if count > 0 {
		return nil
	}

	configs := []models.ExamConfig{
		{
			Name:                 "标准测试",
			DurationMinutes:      30,
			AllowEarlySubmit:     true,
			ForceSubmitOnTimeout: false,
			IsDefault:            true,
		},
		{
			Name:                 "快速练习",
			DurationMinutes:      10,
			AllowEarlySubmit:     true,
			ForceSubmitOnTimeout: false,
			IsDefault:            false,
		},
		{
			Name:                 "限时挑战",
			DurationMinutes:      5,
			AllowEarlySubmit:     true,
			ForceSubmitOnTimeout: false,
			IsDefault:            false,
		},
	}

	for _, config := range configs {
		if err := db.Create(&config).Error; err != nil {
			return fmt.Errorf("seed exam config %s: %w", config.Name, err)
		}
	}
	return nil
}
