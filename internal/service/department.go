package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/iiincognito/org-structure/internal/models"
	"github.com/iiincognito/org-structure/internal/repository"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

var (
	ErrDepartmentNotFound    = errors.New("подразделение не найдено")
	ErrInvalidName           = errors.New("название подразделения некорректно")
	ErrDuplicateNameInParent = errors.New("подразделение с таким названием уже существует у этого родителя")
	ErrSelfAsParent          = errors.New("нельзя сделать подразделение родителем самого себя")
	ErrCycleInHierarchy      = errors.New("перемещение создаст цикл в иерархии")
	ErrInvalidReassignTarget = errors.New("целевое подразделение для переназначения не указано или не существует")
	ErrInvalidDeleteMode     = errors.New("неверный режим удаления")
	ErrInvalidFullName       = errors.New("ФИО сотрудника некорректно (пустое или слишком длинное)")
	ErrInvalidPosition       = errors.New("должность некорректна (пустая или слишком длинная)")
	ErrInvalidHiredAt        = errors.New("некорректный формат даты приёма на работу")
)

type DepartmentService interface {
	CreateDepartment(ctx context.Context, name string, parentID *uint) (*models.Department, error)
	GetDepartment(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error)
	UpdateDepartment(ctx context.Context, id uint, name *string, parentID *uint) (*models.Department, error)
	DeleteDepartment(ctx context.Context, id uint, mode string, reassignTo *uint) error

	CreateEmployee(ctx context.Context, deptID uint, fullName, position string, hiredAt *string) (*models.Employee, error)
}

type departmentService struct {
	repo repository.DepartmentRepository
}

func NewDepartmentService(repo repository.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

// CreateDepartment
func (s *departmentService) CreateDepartment(ctx context.Context, name string, parentID *uint) (*models.Department, error) {
	name = trimAndValidateName(name)
	if name == "" {
		return nil, ErrInvalidName
	}

	dept := &models.Department{
		Name:     name,
		ParentID: parentID,
	}

	if err := s.repo.CreateDepartment(ctx, dept); err != nil {
		// Здесь можно добавить обработку уникальности из БД
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateNameInParent
		}
		return nil, err
	}

	return dept, nil
}

// GetDepartment
func (s *departmentService) GetDepartment(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	dept, err := s.repo.GetDepartmentWithDepth(ctx, id, depth, includeEmployees)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}

	// Можно добавить сортировку сотрудников, если нужно
	if includeEmployees && len(dept.Employees) > 0 {
		sort.Slice(dept.Employees, func(i, j int) bool {
			return dept.Employees[i].CreatedAt.Before(dept.Employees[j].CreatedAt)
		})
	}

	return dept, nil
}

// UpdateDepartment
func (s *departmentService) UpdateDepartment(ctx context.Context, id uint, name *string, parentID *uint) (*models.Department, error) {
	var updatedName *string
	if name != nil {
		n := trimAndValidateName(*name)
		if n == "" {
			return nil, ErrInvalidName
		}
		updatedName = &n
	}

	dept, err := s.repo.UpdateDepartment(ctx, id, updatedName, parentID)
	if err != nil {
		if errors.Is(err, ErrSelfAsParent) || errors.Is(err, ErrCycleInHierarchy) {
			return nil, err // уже бизнес-ошибки
		}
		return nil, err
	}

	return dept, nil
}

// DeleteDepartment
func (s *departmentService) DeleteDepartment(ctx context.Context, id uint, mode string, reassignTo *uint) error {
	mode = strings.ToLower(mode)

	switch mode {
	case "cascade":
		return s.repo.DeleteDepartment(ctx, id, "cascade", nil)
	case "reassign":
		if reassignTo == nil {
			return ErrInvalidReassignTarget
		}
		return s.repo.DeleteDepartment(ctx, id, "reassign", reassignTo)
	default:
		return ErrInvalidDeleteMode
	}
}

func (s *departmentService) CreateEmployee(
	ctx context.Context,
	deptID uint,
	fullName string,
	position string,
	hiredAt *string,
) (*models.Employee, error) {
	// 1. Валидация входных данных
	fullName = strings.TrimSpace(fullName)
	if fullName == "" || len(fullName) > 200 {
		return nil, ErrInvalidFullName
	}

	position = strings.TrimSpace(position)
	if position == "" || len(position) > 200 {
		return nil, ErrInvalidPosition
	}

	// 2. Проверяем существование подразделения

	_, err := s.repo.GetDepartmentWithDepth(ctx, deptID, 1, false)
	if err != nil {
		if errors.Is(err, ErrDepartmentNotFound) {
			return nil, fmt.Errorf("отдел с id %d не найден: %w", deptID, ErrDepartmentNotFound)
		}
		return nil, fmt.Errorf("ошибка проверки отдела: %w", err)
	}

	// 3. Парсим hiredAt, если передан
	var hiredAtTime *time.Time
	if hiredAt != nil && *hiredAt != "" {
		parsed, err := time.Parse("2006-01-02", *hiredAt) // формат YYYY-MM-DD
		if err != nil {
			return nil, fmt.Errorf("некорректный формат hired_at (ожидается YYYY-MM-DD): %w", err)
		}
		hiredAtTime = &parsed
	}

	// 4. Создаём модель
	employee := &models.Employee{
		DepartmentID: deptID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAtTime,
	}

	// 5. Сохраняем через репозиторий
	if err := s.repo.CreateEmployee(ctx, employee); err != nil {
		return nil, fmt.Errorf("ошибка создания сотрудника: %w", err)
	}

	return employee, nil
}

func trimAndValidateName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 200 {
		return ""
	}
	return name
}

func isUniqueConstraintError(err error) bool {

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return true
	}
	return false
}
