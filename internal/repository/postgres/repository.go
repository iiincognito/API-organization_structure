package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/iiincognito/org-structure/internal/models"
	"github.com/iiincognito/org-structure/internal/repository"

	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) repository.DepartmentRepository {
	return &postgresRepository{db: db}
}

// ПОДРАЗДЕЛЕНИЯ

func (r *postgresRepository) CreateDepartment(ctx context.Context, dept *models.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *postgresRepository) GetDepartmentWithDepth(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}

	var dept models.Department
	query := r.db.WithContext(ctx).Preload("Parent")

	// Рекурсивно загружаем детей до нужной глубины
	for i := 1; i <= depth; i++ {
		preloadPath := "Children"
		for j := 1; j < i; j++ {
			preloadPath += ".Children"
		}
		query = query.Preload(preloadPath)
	}

	if includeEmployees {
		query = query.Preload("Employees")
	}

	err := query.First(&dept, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("department not found")
		}
		return nil, err
	}
	return &dept, nil
}

func (r *postgresRepository) UpdateDepartment(ctx context.Context, id uint, name *string, parentID *uint) (*models.Department, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var dept models.Department
	if err := tx.First(&dept, id).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if parentID != nil {
		// Проверка на цикл
		if *parentID == id {
			return nil, errors.New("cannot set self as parent")
		}
		isDescendant, _ := r.IsDescendant(ctx, *parentID, id)
		if isDescendant {
			return nil, errors.New("cannot create cycle in department tree")
		}
		updates["parent_id"] = parentID
	}

	if err := tx.Model(&dept).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Возвращаем обновлённый объект
	return r.GetDepartmentWithDepth(ctx, id, 1, false)
}

func (r *postgresRepository) DeleteDepartment(ctx context.Context, id uint, mode string, reassignTo *uint) error {
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	switch mode {
	case "cascade":

		if err := tx.Delete(&models.Department{}, id).Error; err != nil {
			return err
		}

	case "reassign":
		if reassignTo == nil {
			return errors.New("reassign_to_department_id is required for reassign mode")
		}
		// Переводим сотрудников
		if err := tx.Model(&models.Employee{}).
			Where("department_id = ?", id).
			Update("department_id", reassignTo).Error; err != nil {
			return err
		}
		// Удаляем подразделение (дочерние удалятся по CASCADE)
		if err := tx.Delete(&models.Department{}, id).Error; err != nil {
			return err
		}

	default:
		return errors.New("invalid delete mode")
	}

	return tx.Commit().Error
}

//  СОТРУДНИКИ

func (r *postgresRepository) CreateEmployee(ctx context.Context, emp *models.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

//  ВСПОМОГАТЕЛЬНЫЕ

func (r *postgresRepository) IsDescendant(ctx context.Context, parentID, childID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE tree AS (
			SELECT id, parent_id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id, d.parent_id FROM departments d
			JOIN tree t ON d.parent_id = t.id
		)
		SELECT COUNT(*) FROM tree WHERE id = ?
	`, parentID, childID).Scan(&count).Error

	return count > 0, err
}
