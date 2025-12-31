package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	// ExecDir 执行程序所处目录
	ExecDir string

	// DataDir 数据目录
	DataDir string

	// BinDir 其他执行程序目录
	BinDir string

	// ConfigDir 配置文件存储目录
	ConfigDir string

	// LockFile 防止程序多开 锁文件位置
	LockFile string
)

func init() {
	// 初始化路径
	if err := initPaths(); err != nil {
		log.Fatalf("初始化路径失败: %v", err)
	}
}

// initPaths 初始化所有路径
func initPaths() error {
	var err error

	// 执行程序所处目录
	ExecDir, err = getExecutableDir()
	if err != nil {
		return fmt.Errorf("获取执行程序目录失败: %w", err)
	}

	// data目录
	DataDir = filepath.Join(ExecDir, "data")

	// bin目录
	BinDir = filepath.Join(ExecDir, "bin")

	// 配置文件目录
	ConfigDir = filepath.Join(ExecDir, "config")

	// 防止程序多开锁文件
	LockFile = filepath.Join(ExecDir, "run.lock")

	// 创建 restic 目录
	_ = os.MkdirAll(filepath.Join(DataDir, "/restic/"), 0700)

	// 创建 bin 目录
	_ = os.MkdirAll(BinDir, 0700)

	// 创建 config 目录
	_ = os.MkdirAll(ConfigDir, 0700)

	return nil
}

// getExecutableDir 获取执行程序所在目录
func getExecutableDir() (string, error) {
	// 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// 解析符号链接
	if linked, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = linked
	}

	// 返回目录部分
	return filepath.Dir(execPath), nil
}
