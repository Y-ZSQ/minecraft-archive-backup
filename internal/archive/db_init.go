package archive

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"minecraft-archive-backup/model/dto/database"
	"minecraft-archive-backup/pkg/etc/core"
	"path/filepath"
)

var (
	DB *gorm.DB
)

func init() {
	var (
		err    error
		dbPath = filepath.Join(core.DataDir, "archive.sqlite")
	)

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	if err = DB.AutoMigrate(&database.Archive{}, database.BackupRecord{}); err != nil {
		log.Fatalf("数据库表创建失败: %v", err)
	}

	if err := RefreshArchiveCache(); err != nil {
		log.Fatal(err)
	}
}
