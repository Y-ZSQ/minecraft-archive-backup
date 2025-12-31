package database

import (
	"time"
)

// BackupRecord 存档的历史备份记录
type BackupRecord struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ArchiveID uint `gorm:"index;constraint:OnDelete:CASCADE"`
	Archive   Archive
	SnapShot  string `gorm:"unique;size:64"`
	Comment   string
}
