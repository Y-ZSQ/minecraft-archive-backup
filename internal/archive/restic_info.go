package archive

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type RawData struct {
	TotalSize              int     `json:"total_size"`
	TotalUncompressedSize  int     `json:"total_uncompressed_size"`
	CompressionRatio       float64 `json:"compression_ratio"`
	CompressionProgress    int     `json:"compression_progress"`
	CompressionSpaceSaving float64 `json:"compression_space_saving"`
	TotalBlobCount         int     `json:"total_blob_count"`
	SnapshotsCount         int     `json:"snapshots_count"`
}

// ResticRawData 快照实际占用的大小
func ResticRawData(snapshots string) (*RawData, error) {
	var data = &RawData{}

	if len(snapshots) == 0 {
		return data, errors.New("至少需要指定一个快照")
	}

	// 构建命令参数
	args := []string{"stats"}
	args = append(args, "--json")
	args = append(args, "--mode", "raw-data")

	cmd := NewResticCmd(exec.Command("restic", args...))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return data, fmt.Errorf("获取快照实际占用大小失败: %v\n输出: %s", err, output)
	}

	// 解析输出 并且返回
	err = json.Unmarshal(output, data)

	return data, err
}

type RestoreSize struct {
	TotalSize      int `json:"total_size"`
	TotalFileCount int `json:"total_file_count"`
	SnapshotsCount int `json:"snapshots_count"`
}

// ResticRestoreSize 增量备份 和 压缩后 占用的空间
func ResticRestoreSize(snapshots string) (*RestoreSize, error) {
	var data = &RestoreSize{}

	if len(snapshots) == 0 {
		return data, errors.New("至少需要指定一个快照")
	}

	// 构建命令参数
	args := []string{"stats"}
	args = append(args, "--json")
	args = append(args, "--mode", "restore-size")

	cmd := NewResticCmd(exec.Command("restic", args...))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return data, fmt.Errorf("获取快照真实占用失败: %v\n输出: %s", err, output)
	}

	// 解析输出 并且返回
	err = json.Unmarshal(output, data)

	return data, err
}
