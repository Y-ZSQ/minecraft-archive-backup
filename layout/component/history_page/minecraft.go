package history_page

import (
	"fmt"
	"github.com/gofrs/flock"
	"os"
	"strings"
)

// IsWorldInUse 检查Minecraft世界是否正在被游玩
// lockFilePath: session.lock文件的完整路径
// 返回: 是否正在使用, 错误信息
func IsWorldInUse(lockFilePath string) (bool, error) {
	// 1. 检查文件是否存在
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		return false, nil
	}

	// 2. 检查旧版本 (通过文件内容)
	// 旧版本的 session.lock 并没有上锁 并且每次进入存档 都会有一个新的 16位 16进制字符串
	//oldVersionInUse, err := checkOldVersion(lockFilePath)
	//if err == nil {
	//	return oldVersionInUse, nil
	//}

	// 3. 检查新版本（通过文件锁）
	return checkNewVersion(lockFilePath)
}

// checkNewVersion 检查新版本Minecraft
func checkNewVersion(lockFilePath string) (bool, error) {
	fileLock := flock.New(lockFilePath)

	// 尝试加锁，设置很短的超时时间
	locked, err := fileLock.TryLock()
	if err != nil {
		return false, err
	}

	if locked {
		// 能加锁，说明文件没有被使用
		fileLock.Unlock()
		return false, nil
	}

	// 不能加锁，说明文件被占用
	return true, nil
}

// checkOldVersion 检查旧版本Minecraft
func checkOldVersion(lockFilePath string) (bool, error) {
	content, err := os.ReadFile(lockFilePath)
	if err != nil {
		return false, err
	}

	// 转换为十六进制字符串
	hexStr := fmt.Sprintf("%x", content)
	return isOldVersionInUse(hexStr), nil
}

// isLockError 判断是否是文件锁定错误
func isLockError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "The process cannot access") ||
		strings.Contains(errStr, "being used") ||
		strings.Contains(errStr, "access is denied")
}

// isOldVersionInUse 判断旧版本世界是否在使用
func isOldVersionInUse(hexStr string) bool {
	hexStr = strings.TrimSpace(hexStr)

	// 旧版本session.lock是16位十六进制
	if len(hexStr) != 16 {
		return false
	}

	// 检查是否是有效的十六进制
	for _, c := range hexStr {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return false
}
