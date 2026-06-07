package service

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

type KnowledgePointService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewKnowledgePointService(db *gorm.DB, log *slog.Logger) *KnowledgePointService {
	return &KnowledgePointService{db: db, log: log}
}

func (s *KnowledgePointService) ListKnowledgePoints() ([]models.KnowledgePoint, error) {
	var kps []models.KnowledgePoint
	if err := s.db.Order("sort asc, id asc").Find(&kps).Error; err != nil {
		return nil, fmt.Errorf("list knowledge points: %w", err)
	}
	return kps, nil
}

func (s *KnowledgePointService) GetKnowledgePoint(id uint) (*models.KnowledgePoint, error) {
	var kp models.KnowledgePoint
	if err := s.db.First(&kp, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrKnowledgePointNotFound
		}
		return nil, fmt.Errorf("get knowledge point: %w", err)
	}
	return &kp, nil
}

func (s *KnowledgePointService) GetOrCreateKnowledgePoint(name string) (*models.KnowledgePoint, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrKnowledgePointNameEmpty
	}

	var kp models.KnowledgePoint
	err := s.db.Where("name = ?", name).First(&kp).Error
	if err == nil {
		return &kp, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find knowledge point: %w", err)
	}

	kp = models.KnowledgePoint{Name: name}
	if err := s.db.Create(&kp).Error; err != nil {
		return nil, fmt.Errorf("create knowledge point: %w", err)
	}
	s.log.Info("knowledge point created", "id", kp.ID, "name", kp.Name)
	return &kp, nil
}

func (s *KnowledgePointService) CreateKnowledgePoint(name string, sort int) (*models.KnowledgePoint, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrKnowledgePointNameEmpty
	}

	var existing models.KnowledgePoint
	err := s.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, ErrKnowledgePointExists
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("check knowledge point exists: %w", err)
	}

	kp := models.KnowledgePoint{
		Name: name,
		Sort: sort,
	}
	if err := s.db.Create(&kp).Error; err != nil {
		return nil, fmt.Errorf("create knowledge point: %w", err)
	}
	s.log.Info("knowledge point created", "id", kp.ID, "name", kp.Name)
	return &kp, nil
}

func (s *KnowledgePointService) UpdateKnowledgePoint(id uint, name string, sort int) (*models.KnowledgePoint, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrKnowledgePointNameEmpty
	}

	var kp models.KnowledgePoint
	if err := s.db.First(&kp, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrKnowledgePointNotFound
		}
		return nil, fmt.Errorf("find knowledge point: %w", err)
	}

	var existing models.KnowledgePoint
	err := s.db.Where("name = ? AND id != ?", name, id).First(&existing).Error
	if err == nil {
		return nil, ErrKnowledgePointExists
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("check knowledge point exists: %w", err)
	}

	kp.Name = name
	kp.Sort = sort
	if err := s.db.Save(&kp).Error; err != nil {
		return nil, fmt.Errorf("update knowledge point: %w", err)
	}
	s.log.Info("knowledge point updated", "id", kp.ID, "name", kp.Name)
	return &kp, nil
}

func (s *KnowledgePointService) DeleteKnowledgePoint(id uint) error {
	res := s.db.Delete(&models.KnowledgePoint{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete knowledge point: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrKnowledgePointNotFound
	}

	if err := s.db.Where("knowledge_point_id = ?", id).Delete(&models.QuestionKnowledgePoint{}).Error; err != nil {
		return fmt.Errorf("delete question knowledge point relations: %w", err)
	}

	s.log.Info("knowledge point deleted", "id", id)
	return nil
}

func toKnowledgePointInfos(kps []models.KnowledgePoint) []dto.KnowledgePointInfo {
	result := make([]dto.KnowledgePointInfo, 0, len(kps))
	for _, kp := range kps {
		result = append(result, dto.KnowledgePointInfo{
			ID:   kp.ID,
			Name: kp.Name,
		})
	}
	return result
}
