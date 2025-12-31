package archive

import (
	"fmt"
	"minecraft-archive-backup/model/dto/database"
	"os/exec"
	"strings"
)

// ResticRestore 回档到指定的快照id时
func ResticRestore(archive *database.Archive, record *database.BackupRecord) <-chan *BackupMessage {
	// 构建命令参数
	args := []string{"restore", "--json"}
	args = append(args, fmt.Sprintf("%s:%s",
		record.SnapShot,
		ConvertWindowsToUnixPath(archive.Path)))
	args = append(args, "--target", archive.Path)

	cmd := NewResticCmd(exec.Command("restic", args...))

	return executeResticCommand(cmd)
}

// ConvertWindowsToUnixPath 将Windows路径转换为Unix格式
func ConvertWindowsToUnixPath(windowsPath string) string {
	// 将反斜杠替换为正斜杠
	unixPath := strings.ReplaceAll(windowsPath, "\\", "/")

	// 处理盘符（如 C:\ 转换为 /C/）
	if len(unixPath) >= 2 && unixPath[1] == ':' {
		// 如果是盘符绝对路径
		if unixPath[2] == '/' {
			// 如 C:/path -> /C/path
			return fmt.Sprintf("/%c%s", unixPath[0], unixPath[2:])
		} else if len(unixPath) == 2 {
			// 如 C: -> /C
			return fmt.Sprintf("/%c", unixPath[0])
		}
	}

	return unixPath
}
