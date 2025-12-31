package archive_info_page

import (
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
)

// IsValidPathFormat 检查是否是有效的路径格式
func IsValidPathFormat(path string) bool {
	if path == "" {
		return false
	}

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 基本检查
	if strings.Contains(cleanPath, "\x00") {
		return false
	}

	// Windows 特定检查
	if runtime.GOOS == "windows" {
		return isValidWindowsPath(cleanPath)
	}

	return isValidUnixPath(cleanPath)
}

// isValidWindowsPath Windows 路径验证
func isValidWindowsPath(path string) bool {
	// Windows 非法字符（除了驱动器后的 :）
	illegalChars := `<>"|?*`

	// 检查非法字符（需要特殊处理 :）
	for i := 0; i < len(path); {
		r, size := utf8.DecodeRuneInString(path[i:])

		// 检查是否是非法字符
		if strings.ContainsRune(illegalChars, r) {
			return false
		}

		// 特殊处理 : 字符
		if r == ':' {
			// 只有驱动器盘符后的 : 是合法的
			// 格式应该是 X: 或 X:\
			if i != 1 { // 不是第二个字符
				return false
			}
			// 检查前面是否是字母
			if i > 0 {
				prevRune, _ := utf8.DecodeRuneInString(path[:i])
				if !isWindowsDriveLetter(prevRune) {
					return false
				}
			}
		}

		i += size
	}

	// 检查保留设备名
	base := filepath.Base(path)
	if isWindowsReservedName(strings.ToLower(base)) {
		return false
	}

	// 检查路径末尾的点或空格
	if strings.HasSuffix(path, ".") || strings.HasSuffix(strings.TrimRight(path, " "), ".") {
		return false
	}

	return true
}

// isWindowsDriveLetter 检查是否是 Windows 驱动器字母
func isWindowsDriveLetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// isWindowsReservedName 检查是否是 Windows 保留名
func isWindowsReservedName(name string) bool {
	reserved := []string{
		"con", "prn", "aux", "nul",
		"com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9",
		"lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9",
	}

	// 去掉扩展名
	if dot := strings.LastIndex(name, "."); dot > 0 {
		name = name[:dot]
	}

	for _, reservedName := range reserved {
		if strings.EqualFold(name, reservedName) {
			return true
		}
	}
	return false
}

// isValidUnixPath Unix/Linux/Mac 路径验证
func isValidUnixPath(path string) bool {
	return !strings.ContainsRune(path, 0)
}
