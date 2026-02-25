package models

import "time"

type Department struct {
	ID        uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string       `gorm:"size:200;not null;index" json:"name"` // длина 1-200 + индекс
	ParentID  *uint        `gorm:"index" json:"parent_id"`              // NULL = корневой
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID;constraint:OnDelete:SET NULL" json:"employees,omitempty"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Parent    *Department  `gorm:"foreignKey:ParentID" json:"-"`
}

func (Department) TableName() string {
	return "departments"
}
