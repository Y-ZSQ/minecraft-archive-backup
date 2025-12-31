package archive

import (
	"errors"
	"gorm.io/gorm"
	"minecraft-archive-backup/model/dto/database"
)

// CreateBackupRecord 创建一个备份记录
func CreateBackupRecord(backupRecord *database.BackupRecord) error {
	// 检查关联的存档是否存在
	var archive database.Archive
	result := DB.First(&archive, backupRecord.ArchiveID)
	if result.Error != nil {
		return errors.New("关联的存档不存在")
	}

	result = DB.Create(backupRecord)
	return result.Error
}

// UpdateBackupRecord 更新一个备份记录
func UpdateBackupRecord(backupRecord *database.BackupRecord) error {
	// 检查备份记录是否存在
	var existingRecord database.BackupRecord
	result := DB.First(&existingRecord, backupRecord.ID)
	if result.Error != nil {
		return errors.New("备份记录不存在")
	}

	result = DB.Save(backupRecord)
	return result.Error
}

// DeleteBackupRecord 删除指定备份记录
func DeleteBackupRecord(id uint) error {
	result := DB.Delete(&database.BackupRecord{}, id)
	return result.Error
}

// GetBackupRecordsByArchiveID 查询指定存档ID下的所有备份记录
func GetBackupRecordsByArchiveID(archiveID uint) ([]database.BackupRecord, error) {
	var backupRecords []database.BackupRecord
	result := DB.Where("archive_id = ?", archiveID).Find(&backupRecords)
	return backupRecords, result.Error
}

// GetBackupRecordByID 通过备份记录ID查询指定备份记录
func GetBackupRecordByID(backupID uint) (*database.BackupRecord, error) {
	var backupRecord database.BackupRecord
	result := DB.Preload("Archive").First(&backupRecord, backupID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &backupRecord, result.Error
}
