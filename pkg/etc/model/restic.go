package model

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	etc "minecraft-archive-backup/pkg/etc/core"
	"path/filepath"
)

type Restic struct {
	Password string
}

func (r *Restic) Key() string {
	return "Restic"
}

func (r *Restic) FilePath() string {
	return filepath.Join(etc.ConfigDir, "restic.yaml")
}

func (r *Restic) DefaultValueMap() map[string]any {
	return map[string]any{
		"password": GeneratePassword(),
	}
}

func (r *Restic) WatchFun() func(fsnotify.Event) {
	return func(event fsnotify.Event) {

	}
}

func GeneratePassword() string {
	// 生成UUID
	id := uuid.New().String()

	// SHA256加密
	hash := sha256.Sum256([]byte(id))

	// 转为16进制字符串
	password := hex.EncodeToString(hash[:])

	return password
}
