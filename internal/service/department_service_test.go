package service

import (
	"context"
	"errors"
	"github.com/iiincognito/org-structure/internal/models"
	"github.com/iiincognito/org-structure/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDepartmentService_CreateEmployee_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.DepartmentRepository)
	svc := NewDepartmentService(mockRepo)

	ctx := context.Background()
	deptID := uint(42)
	fullName := "Петров Сергей Александрович"
	position := "Ведущий разработчик"
	hiredAtStr := "2025-03-10"

	expectedHiredAt, err := time.Parse("2006-01-02", hiredAtStr)
	require.NoError(t, err) // лучше проверять парсинг

	// Настраиваем мок на РЕАЛЬНЫЙ метод
	mockRepo.
		On("GetDepartmentWithDepth", ctx, deptID, 1, false).
		Return(&models.Department{ID: deptID}, nil).
		Once()

	mockRepo.
		On("CreateEmployee", ctx, mock.MatchedBy(func(emp *models.Employee) bool {
			return emp != nil &&
				emp.DepartmentID == deptID &&
				emp.FullName == fullName &&
				emp.Position == position &&
				emp.HiredAt != nil &&
				emp.HiredAt.Equal(expectedHiredAt)
		})).
		Return(nil).
		Once()

	// Act
	created, err := svc.CreateEmployee(ctx, deptID, fullName, position, &hiredAtStr)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, created)

	assert.Equal(t, deptID, created.DepartmentID)
	assert.Equal(t, fullName, created.FullName)
	assert.Equal(t, position, created.Position)
	assert.NotNil(t, created.HiredAt)
	assert.Equal(t, expectedHiredAt, *created.HiredAt)

	mockRepo.AssertExpectations(t)
}

func TestDepartmentService_CreateEmployee_DepartmentNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.DepartmentRepository)
	svc := NewDepartmentService(mockRepo)

	ctx := context.Background()
	deptID := uint(999)

	// Настраиваем мок на реальный метод GetDepartmentWithDepth
	mockRepo.
		On("GetDepartmentWithDepth", ctx, deptID, 1, false).
		Return(nil, errors.New("department not found")).
		Once()

	// Act
	emp, err := svc.CreateEmployee(ctx, deptID, "Тестов Т.Т.", "Тестер", nil)

	// Assert
	require.Error(t, err)
	assert.Nil(t, emp)
	assert.Contains(t, err.Error(), "ошибка проверки отдела") // или точное сообщение из твоего сервиса

	// Проверяем, что CreateEmployee НЕ был вызван
	mockRepo.AssertNotCalled(t, "CreateEmployee")
	mockRepo.AssertExpectations(t)
}
