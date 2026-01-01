package archive

import (
	"errors"
	"fmt"
	"os/exec"
)

// ResticForget 删除指定的快照
func ResticForget(snapshots ...string) error {
	if len(snapshots) == 0 {
		return errors.New("至少需要指定一个快照")
	}

	// 构建命令参数
	args := []string{"forget"}
	args = append(args, "--json")
	args = append(args, "--prune")
	args = append(args, snapshots...)

	cmd := NewResticCmd(exec.Command("restic", args...))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除快照失败: %v\n输出: %s", err, output)
	}

	return nil
}
