package models

import (
	"time"
)

type Employee struct {
	ID           uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentID uint        `gorm:"not null;index" json:"department_id"` // FK
	FullName     string      `gorm:"size:200;not null;index" json:"full_name"`
	Position     string      `gorm:"size:200;not null" json:"position"`
	HiredAt      *time.Time  `gorm:"index" json:"hired_at,omitempty"` // мб NULL
	CreatedAt    time.Time   `gorm:"autoCreateTime" json:"created_at"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"-"`
}

func (Employee) TableName() string {
	return "employees"
}
