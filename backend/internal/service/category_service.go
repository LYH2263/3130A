package service

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"label3130/backend/internal/models"
)

type CategoryService struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewCategoryService(db *gorm.DB, log *slog.Logger) *CategoryService {
	return &CategoryService{db: db, log: log}
}

func (s *CategoryService) ListTree() ([]models.Category, error) {
	var allCategories []models.Category
	if err := s.db.Order("sort asc, id asc").Find(&allCategories).Error; err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	type catCount struct {
		CategoryID uint
		Count      int64
	}
	var counts []catCount
	if err := s.db.Model(&models.Question{}).
		Select("category_id as category_id, count(*) as count").
		Where("category_id IS NOT NULL").
		Group("category_id").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("count category questions: %w", err)
	}

	countMap := make(map[uint]int64)
	for _, c := range counts {
		countMap[c.CategoryID] = c.Count
	}

	tree := buildTree(allCategories, nil)
	fillQuestionCount(tree, countMap)

	return tree, nil
}

func fillQuestionCount(categories []models.Category, countMap map[uint]int64) int64 {
	var total int64
	for i := range categories {
		directCount := countMap[categories[i].ID]
		childCount := fillQuestionCount(categories[i].Children, countMap)
		categories[i].QuestionCount = int(directCount + childCount)
		total += directCount + childCount
	}
	return total
}

func buildTree(categories []models.Category, parentID *uint) []models.Category {
	result := make([]models.Category, 0)
	for _, cat := range categories {
		var isMatch bool
		if parentID == nil {
			isMatch = cat.ParentID == nil
		} else {
			isMatch = cat.ParentID != nil && *cat.ParentID == *parentID
		}
		if isMatch {
			children := buildTree(categories, &cat.ID)
			cat.Children = children
			result = append(result, cat)
		}
	}
	return result
}

func (s *CategoryService) GetCategory(id uint) (*models.Category, error) {
	var category models.Category
	if err := s.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &category, nil
}

func (s *CategoryService) GetDescendantIDs(id uint) ([]uint, error) {
	var ids []uint
	ids = append(ids, id)

	var children []models.Category
	if err := s.db.Where("parent_id = ?", id).Find(&children).Error; err != nil {
		return nil, fmt.Errorf("get children: %w", err)
	}

	for _, child := range children {
		childIDs, err := s.GetDescendantIDs(child.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, childIDs...)
	}

	return ids, nil
}

func (s *CategoryService) CreateCategory(name string, parentID *uint, sort int) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrCategoryNameEmpty
	}

	if parentID != nil {
		var parent models.Category
		if err := s.db.First(&parent, *parentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, ErrCategoryParentInvalid
			}
			return nil, fmt.Errorf("check parent: %w", err)
		}
	}

	category := models.Category{
		Name:     name,
		ParentID: parentID,
		Sort:     sort,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	s.log.Info("category created", "categoryID", category.ID, "name", category.Name)
	return &category, nil
}

func (s *CategoryService) UpdateCategory(id uint, name string, parentID *uint, sort int) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrCategoryNameEmpty
	}

	var category models.Category
	if err := s.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("find category: %w", err)
	}

	if parentID != nil {
		if *parentID == id {
			return nil, ErrCategoryParentInvalid
		}
		var parent models.Category
		if err := s.db.First(&parent, *parentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, ErrCategoryParentInvalid
			}
			return nil, fmt.Errorf("check parent: %w", err)
		}
	}

	category.Name = name
	category.ParentID = parentID
	category.Sort = sort

	if err := s.db.Save(&category).Error; err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}

	s.log.Info("category updated", "categoryID", category.ID)
	return &category, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	result := s.db.Delete(&models.Category{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	s.log.Info("category deleted", "categoryID", id)
	return nil
}
