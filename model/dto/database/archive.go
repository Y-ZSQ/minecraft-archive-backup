package database

import (
	"time"
)

// Archive 存档的信息
type Archive struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `gorm:"unique;not null"` // 存档的名称 唯一
	Comment   string // 存档的备注 可以为空
	Path      string `gorm:"unique;not null"` // 存档的路径 唯一
}
