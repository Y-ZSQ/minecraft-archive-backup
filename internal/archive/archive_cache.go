package archive

import (
	"errors"
	"fmt"
	"maps"
	"minecraft-archive-backup/model/dto/database"
	"sync"
)

// 自定义错误类型
var (
	ErrCacheNotFound = errors.New("缓存不存在")
	ErrInvalidKey    = errors.New("无效的键")
	ErrInvalidValue  = errors.New("无效的值")
)

// 缓存实例
var (
	cache   = make(map[uint]*database.Archive)
	cacheMu sync.RWMutex
)

// RefreshArchiveCache 重新读取数据库中的存档 并载入到内存中
func RefreshArchiveCache() error {
	var archives, err = GetAllArchives()
	if err != nil {
		return fmt.Errorf("加载所有存档到缓存失败: %v", err)
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	for _, archive := range archives {
		cache[archive.ID] = &archive
	}

	return nil
}

// LoadAllArchiveCache 在缓存中加载所有的存档
func LoadAllArchiveCache() map[uint]*database.Archive {
	return maps.Clone(cache)
}

// LoadArchiveCache 加载缓存
func LoadArchiveCache(key uint) (*database.Archive, error) {
	if key == 0 {
		return nil, ErrInvalidKey
	}

	cacheMu.RLock()
	archive, exists := cache[key]
	cacheMu.RUnlock()

	if !exists {
		return nil, ErrCacheNotFound
	}

	return archive, nil
}

// StoreArchiveCache 存储到缓存
func StoreArchiveCache(archive *database.Archive) error {
	if archive == nil {
		return ErrInvalidValue
	}
	if archive.ID == 0 {
		return ErrInvalidKey
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	cache[archive.ID] = archive
	return nil
}

// DeleteArchiveCache 删除缓存
func DeleteArchiveCache(key uint) error {
	if key == 0 {
		return ErrInvalidKey
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	delete(cache, key)
	return nil
}

// ExistsArchiveCache 检查缓存是否存在
func ExistsArchiveCache(key uint) bool {
	if key == 0 {
		return false
	}

	cacheMu.RLock()
	_, exists := cache[key]
	cacheMu.RUnlock()

	return exists
}

// GetOrCreateArchiveCache 获取缓存，不存在则创建
// 如果打算直接创建一个新的存档 以及 他的缓存 那么 请将 key == 0
func GetOrCreateArchiveCache(key uint, createFunc func() (*database.Archive, error)) (*database.Archive, error) {

	switch key {
	case 0: // 直接创建
		break
	default: // 正常的读取
		// 尝试读取
		cacheMu.RLock()
		if archive, exists := cache[key]; exists {
			cacheMu.RUnlock()
			return archive, nil
		}
		cacheMu.RUnlock()
	}

	// 创建新值
	archive, err := createFunc()
	if err != nil {
		return nil, err
	}

	key = archive.ID

	// 获取写锁
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// 再次检查，防止并发创建
	if existing, exists := cache[key]; exists {
		return existing, nil
	}

	cache[key] = archive
	return archive, nil
}

// ClearArchiveCache 清空缓存
func ClearArchiveCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// 重新分配一个新的map，旧的让GC回收
	cache = make(map[uint]*database.Archive)
}
