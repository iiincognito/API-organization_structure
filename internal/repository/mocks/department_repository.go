package mocks

import (
	"context"
	"github.com/iiincognito/org-structure/internal/models"
	"github.com/stretchr/testify/mock"
)

type DepartmentRepository struct {
	mock.Mock
}

func (m *DepartmentRepository) CreateDepartment(ctx context.Context, dept *models.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *DepartmentRepository) GetDepartmentWithDepth(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	args := m.Called(ctx, id, depth, includeEmployees)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

func (m *DepartmentRepository) CreateEmployee(ctx context.Context, emp *models.Employee) error {
	args := m.Called(ctx, emp)
	return args.Error(0)
}

// UpdateDepartment
func (m *DepartmentRepository) UpdateDepartment(
	ctx context.Context,
	id uint,
	name *string,
	parentID *uint,
) (*models.Department, error) {
	args := m.Called(ctx, id, name, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

// DeleteDepartment
func (m *DepartmentRepository) DeleteDepartment(
	ctx context.Context,
	id uint,
	mode string,
	reassignTo *uint,
) error {
	args := m.Called(ctx, id, mode, reassignTo)
	return args.Error(0)
}

func (m *DepartmentRepository) IsDescendant(ctx context.Context, parentID, childID uint) (bool, error) {
	args := m.Called(ctx, parentID, childID)
	return args.Bool(0), args.Error(1)
}
