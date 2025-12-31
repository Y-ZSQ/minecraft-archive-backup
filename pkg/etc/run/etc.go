package run

import (
	etc "minecraft-archive-backup/pkg/etc/core"
	"minecraft-archive-backup/pkg/etc/model"
)

var (
	cfg = etc.NewConfig()
)

func init() {
	_ = cfg.AddConfig(&model.Restic{})
}

// LoadVipers 加载某一个配置
func LoadVipers(key string) (*etc.SafeViper, error) {
	return cfg.LoadVipers(key)
}
