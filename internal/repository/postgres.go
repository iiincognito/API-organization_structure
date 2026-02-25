package repository

import (
	"context"
	"github.com/iiincognito/org-structure/internal/models"
)

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, dept *models.Department) error
	GetDepartmentWithDepth(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error)
	UpdateDepartment(ctx context.Context, id uint, name *string, parentID *uint) (*models.Department, error)
	DeleteDepartment(ctx context.Context, id uint, mode string, reassignTo *uint) error

	CreateEmployee(ctx context.Context, emp *models.Employee) error

	IsDescendant(ctx context.Context, parentID, childID uint) (bool, error)
}
