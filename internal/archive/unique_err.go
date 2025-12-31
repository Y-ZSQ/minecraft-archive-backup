package archive

import (
	"fmt"
	"strings"
)

// FieldNameToChinese 字段名到中文名的映射
var FieldNameToChinese = map[string]string{
	"archives.name": "存档名称",
	"archives.path": "存档路径",
}

// WrapUniqueConstraintError 包装唯一约束错误，返回友好的中文错误
func WrapUniqueConstraintError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// SQLite 错误格式: UNIQUE constraint failed: archives.name
	if strings.Contains(errStr, "UNIQUE constraint failed:") {
		parts := strings.Split(errStr, ": ")
		if len(parts) == 2 {
			field := strings.TrimSpace(parts[1])

			// 查找中文名
			fieldCN, exists := FieldNameToChinese[field]
			if !exists {
				// 如果没有找到映射，使用字段名
				fieldParts := strings.Split(field, ".")
				if len(fieldParts) == 2 {
					fieldCN = fieldParts[1]
				} else {
					fieldCN = field
				}
			}

			return fmt.Errorf("%s已存在，请使用其他%s", fieldCN, fieldCN)
		}
	}

	// 返回原始错误
	return err
}
