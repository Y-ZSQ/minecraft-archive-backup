package core

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config 整体配置对象
type Config struct {
	// lock 确保并发安全
	lock sync.RWMutex
	// vipers 多文件的配置组
	vipers map[string]*SafeViper
	// configInfos 文件的信息存储
	configInfos map[string]Configure
}

// SafeViper 安全的并发配置操作对象
type SafeViper struct {
	// Mu 使用读写锁 确保该局部并发安全
	Mu sync.RWMutex
	// V 具体的配置操作对象
	V *viper.Viper
}

// Configure 对指定配置格式 进行约束
type Configure interface {
	// Key 返回当前配置 key 唯一id
	Key() string
	// FilePath 返回当前配置 所处文件路径
	FilePath() string
	// DefaultValueMap 返回指定默认的 key-value
	// 并且 viper 自带的默认配置 (自带的默认配置 会导致 默认配置无法被后期修改)
	DefaultValueMap() map[string]any
	// WatchFun 监听配置文件变化 并且执行指定的响应通知事件
	// 如果文件配置被修改 那么内存中的配置变量 会自动更新
	// 如果 返回的函数 == nil 那么即使文件被修改 也不会更新内存的配置
	WatchFun() func(fsnotify.Event)
}

// NewConfig 初始化一个配置对象
func NewConfig() *Config {
	return &Config{
		vipers:      make(map[string]*SafeViper),
		configInfos: make(map[string]Configure),
	}
}

// AddConfig 添加一个新的配置项
func (c *Config) AddConfig(newCfg Configure) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 后续用到的变量
	var (
		key      = newCfg.Key()
		filePath = newCfg.FilePath()
	)

	// 判断传入的 key 以及 filePath 是否符合格式
	if key == "" {
		return errors.New("key is empty")
	}
	if filePath == "" {
		return errors.New("filepath is empty")
	}

	// 判断配置是否已经存在
	if _, ok := c.configInfos[key]; ok {
		return errors.New("config is exist")
	}

	// 添加配置信息记录
	c.configInfos[key] = newCfg

	// 添加具体配置到 vipers
	var v = viper.New()

	// 绑定 viper
	v.SetConfigFile(filePath)
	v.SetConfigType(fileExt(filePath))

	// 读取配置信息
	if err := v.ReadInConfig(); err != nil {
		// 判断文件 是否不存在
		if !strings.Contains(err.Error(), "no such file or directory") &&
			!strings.Contains(err.Error(), " cannot find the file") {
			return fmt.Errorf("failed to read config: %v", err)
		} else {
			// 配置文件不存在 创建该文件 并且将默认值配置 导入到 指定文件中
			// /root/.config/user_manage/config/database.yaml

			//设置参数默认值
			var defaultMap = newCfg.DefaultValueMap()
			if defaultMap != nil {
				for key, value := range defaultMap {
					v.Set(key, value)
				}
			}

			// 确保目录存在
			if err := os.MkdirAll(filepath.Dir(filepath.Dir(filePath)), 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %v", err)
			}

			// 写入配置文件
			if err := v.WriteConfigAs(filePath); err != nil {
				return fmt.Errorf("failed to create config file with defaults: %v", err)
			}
		}
	}

	// 设置函数响应事件
	var changeEvent = newCfg.WatchFun()
	if changeEvent != nil {
		v.WatchConfig()
		v.OnConfigChange(changeEvent)
	}

	// 插入新的配置项
	var safeViper = &SafeViper{
		V: v,
	}
	c.vipers[key] = safeViper

	return nil
}

// LoadVipers 加载一个安全的 配置对象
func (c *Config) LoadVipers(key string) (*SafeViper, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// 从 map 中获取指定的 vipers 对象
	if key == "" {
		return nil, errors.New("key is empty")
	}

	// 判断 key 是否存在
	var v, ok = c.vipers[key]
	if !ok {
		return nil, errors.New("key not exist")
	}

	// 加锁 确保竞争并发安全
	v.Mu.RLock()
	defer v.Mu.RUnlock()

	// 获取指定 viper 对象
	return v, nil
}

// Keys 返回当前绑定的所有配置文件key
func (c *Config) Keys() (result []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for key := range c.configInfos {
		result = append(result, key)
	}
	return
}

// fileExt 辅助函数：从文件路径提取后缀名（如 ".yaml" -> "yaml"）
func fileExt(path string) string {
	if pos := strings.LastIndex(path, "."); pos != -1 {
		return path[pos+1:]
	}
	return "yaml"
}
