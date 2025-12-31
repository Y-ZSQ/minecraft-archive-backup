package archive

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	etc "minecraft-archive-backup/pkg/etc/core"
	etcRun "minecraft-archive-backup/pkg/etc/run"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

var resticConfig, _ = etcRun.LoadVipers("Restic")

func NewResticCmd(cmd *exec.Cmd) *exec.Cmd {
	resticConfig.Mu.RLock()
	defer resticConfig.Mu.RUnlock()

	cmd.Path = filepath.Join(etc.BinDir, "restic.exe")

	// 创建目录
	repoPath := filepath.Join(etc.DataDir, "/restic/repo")
	cachePath := filepath.Join(etc.DataDir, "/restic/cache")

	// 设置环境变量
	cmd.Env = []string{
		fmt.Sprintf("RESTIC_REPOSITORY=%s", repoPath),
		fmt.Sprintf("RESTIC_PASSWORD=%s", resticConfig.V.GetString("password")),
		fmt.Sprintf("RESTIC_CACHE_DIR=%s", cachePath),
	}

	// 静默隐藏窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	return cmd
}

// init 初始化 创建仓库和缓存目录
func init() {
	// 1. 检查 Restic 可执行文件是否存在
	resticPath := filepath.Join(etc.BinDir, "restic.exe")
	if _, err := os.Stat(resticPath); os.IsNotExist(err) {
		fmt.Println("Restic 可执行文件不存在")
		return
	}

	// 2. 创建仓库目录和缓存目录
	repoPath := filepath.Join(etc.DataDir, "/restic/repo")
	cachePath := filepath.Join(etc.DataDir, "/restic/cache")

	// 创建目录
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		fmt.Printf("创建仓库目录失败: %v\n", err)
		return
	}
	fmt.Printf("创建仓库目录: %s\n", repoPath)

	if err := os.MkdirAll(cachePath, 0755); err != nil {
		fmt.Printf("创建缓存目录失败: %v\n", err)
		return
	}
	fmt.Printf("创建缓存目录: %s\n", cachePath)

	// 3. 检查仓库是否已初始化
	checkCmd := exec.Command(resticPath, "cat", "config")
	checkCmd = NewResticCmd(checkCmd)

	// 如果检查成功，说明仓库已存在
	if err := checkCmd.Run(); err == nil {
		fmt.Println("Restic 仓库已存在")
		return
	}

	// 4. 初始化仓库
	fmt.Println("正在初始化 Restic 仓库...")
	initCmd := exec.Command(resticPath, "init")
	initCmd = NewResticCmd(initCmd)

	if output, err := initCmd.CombinedOutput(); err != nil {
		errStr := string(output)
		if contains(errStr, "config file already exists") {
			fmt.Println("Restic 仓库已存在")
		} else {
			fmt.Printf("初始化仓库失败: %v\n输出: %s\n", err, errStr)
		}
	} else {
		fmt.Println("Restic 仓库初始化成功")
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

type BackupMessage struct {
	// 通用字段
	MessageType    string  `json:"message_type"`
	SecondsElapsed int     `json:"seconds_elapsed,omitempty"`
	PercentDone    float64 `json:"percent_done,omitempty"`

	// 错误相关字段
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`

	// 状态相关字段
	TotalFiles int   `json:"total_files,omitempty"`
	FilesDone  int   `json:"files_done,omitempty"`
	TotalBytes int64 `json:"total_bytes,omitempty"`
	BytesDone  int64 `json:"bytes_done,omitempty"`

	// 摘要相关字段
	FilesNew            int     `json:"files_new,omitempty"`
	FilesChanged        int     `json:"files_changed,omitempty"`
	FilesUnmodified     int     `json:"files_unmodified,omitempty"`
	DirsNew             int     `json:"dirs_new,omitempty"`
	DirsChanged         int     `json:"dirs_changed,omitempty"`
	DirsUnmodified      int     `json:"dirs_unmodified,omitempty"`
	DataBlobs           int     `json:"data_blobs,omitempty"`
	TreeBlobs           int     `json:"tree_blobs,omitempty"`
	DataAdded           int64   `json:"data_added,omitempty"`
	DataAddedPacked     int64   `json:"data_added_packed,omitempty"`
	TotalFilesProcessed int     `json:"total_files_processed,omitempty"`
	TotalBytesProcessed int64   `json:"total_bytes_processed,omitempty"`
	TotalDuration       float64 `json:"total_duration,omitempty"`

	// 时间字段
	BackupStart string `json:"backup_start,omitempty"`
	BackupEnd   string `json:"backup_end,omitempty"`
	SnapshotID  string `json:"snapshot_id,omitempty"`
}

// 通用的执行函数
func executeResticCommand(cmd *exec.Cmd) <-chan *BackupMessage {
	outputChan := make(chan *BackupMessage, 100)

	go func() {
		defer close(outputChan)

		// 获取 stdout 和 stderr 管道
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			outputChan <- &BackupMessage{
				MessageType: "error",
				Message:     fmt.Sprintf("无法获取标准输出管道: %v", err),
				Code:        1,
			}
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			outputChan <- &BackupMessage{
				MessageType: "error",
				Message:     fmt.Sprintf("无法获取标准错误管道: %v", err),
				Code:        1,
			}
			return
		}

		// 启动命令
		if err := cmd.Start(); err != nil {
			outputChan <- &BackupMessage{
				MessageType: "error",
				Message:     fmt.Sprintf("无法启动命令: %v", err),
				Code:        1,
			}
			return
		}

		// 创建多路读取器
		multiReader := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(multiReader)

		// 实时读取输出
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 尝试解析 JSON
			var msg BackupMessage
			if jsonErr := json.Unmarshal([]byte(line), &msg); jsonErr == nil {
				outputChan <- &msg
			} else {
				// 不是 JSON，作为文本消息发送
				msgType := "info"
				if strings.Contains(strings.ToLower(line), "error") ||
					strings.Contains(strings.ToLower(line), "fatal") ||
					strings.Contains(strings.ToLower(line), "panic") {
					msgType = "error"
				}

				outputChan <- &BackupMessage{
					MessageType: msgType,
					Message:     line,
				}
			}
		}

		// 等待命令结束
		if err := cmd.Wait(); err != nil {
			outputChan <- &BackupMessage{
				MessageType: "error",
				Message:     fmt.Sprintf("命令执行失败: %v", err),
				Code:        1,
			}
		} else {
			outputChan <- &BackupMessage{
				MessageType: "done",
				Message:     "命令完成",
			}
		}
	}()

	return outputChan
}
