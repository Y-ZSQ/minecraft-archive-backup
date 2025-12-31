package archive

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"minecraft-archive-backup/model/dto/database"
)

// CreateArchive 创建一个存档
func CreateArchive(archive *database.Archive) error {
	result := DB.Create(archive)
	return WrapUniqueConstraintError(result.Error)
}

// UpdateArchive 更新一个存档
func UpdateArchive(archive *database.Archive) error {
	result := DB.Save(archive)
	return WrapUniqueConstraintError(result.Error)
}

// DeleteArchive 删除一个存档，
func DeleteArchive(id uint) error {
	// 使用事务确保数据一致性
	return DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("archive_id = ?", id).Delete(&database.BackupRecord{})
		if result.Error != nil {
			return fmt.Errorf("删除备份记录失败: %w", result.Error)
		}

		result = tx.Delete(&database.Archive{}, id)
		if result.Error != nil {
			return fmt.Errorf("删除存档失败: %w", result.Error)
		}

		return nil
	})
}

// GetAllArchives 查询所有的存档
func GetAllArchives() ([]database.Archive, error) {
	var archives []database.Archive
	result := DB.Find(&archives)
	return archives, result.Error
}

// GetArchiveByID 查询指定ID（主键）的存档
func GetArchiveByID(id uint) (*database.Archive, error) {
	var archive database.Archive
	result := DB.First(&archive, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &archive, result.Error
}
