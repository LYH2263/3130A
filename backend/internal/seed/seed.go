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
			Title:       "HTTP 状态码 404 表示？",
			Description: "Web 基础",
			CreatedBy:   1,
			Options: []models.QuestionOption{
				{Content: "服务器内部错误", IsCorrect: false},
				{Content: "资源未找到", IsCorrect: true},
				{Content: "请求成功", IsCorrect: false},
				{Content: "未授权", IsCorrect: false},
			},
		},
		{
			Title:       "Git 用于查看提交历史的命令是？",
			Description: "开发工具",
			CreatedBy:   1,
			Options: []models.QuestionOption{
				{Content: "git push", IsCorrect: false},
				{Content: "git log", IsCorrect: true},
				{Content: "git reset", IsCorrect: false},
				{Content: "git clean", IsCorrect: false},
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
