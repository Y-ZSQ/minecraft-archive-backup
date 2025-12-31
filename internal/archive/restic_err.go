package archive

import (
	"strings"
)

// ResticError 简化的错误类型
type ResticError struct {
	RawOutput string
	Message   string
}

func (e *ResticError) Error() string {
	return e.Message
}

// ParseResticError 提取报错
func ParseResticError(output string) error {
	// 1. 权限错误
	if strings.Contains(output, "Access is denied") {
		return &ResticError{
			RawOutput: output,
			Message:   "无权限读取存档，请使用管理员模式运行程序",
		}
	}

	// 2. 文件不存在错误
	if strings.Contains(output, "does not exist, skipping") &&
		strings.Contains(output, "all source directories/files do not exist") {
		return &ResticError{
			RawOutput: output,
			Message:   "存档目录不存在，请检查路径",
		}
	}

	// 其他错误返回原始错误
	return &ResticError{
		RawOutput: output,
		Message:   "备份失败: " + output,
	}
}
