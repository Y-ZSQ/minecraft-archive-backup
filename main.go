package main

import (
	"errors"
	"github.com/nightlyone/lockfile"
	"log"
	"minecraft-archive-backup/layout/component/home_page"
	etc "minecraft-archive-backup/pkg/etc/core"
)

func main() {
	// 防止程序多开 出现问题
	//var closeLock = tryRun()
	//defer func() {
	//	if err := closeLock(); err != nil {
	//		log.Fatal("释放锁文件失败")
	//	}
	//}()

	// 启动窗口
	home_page.Run()
}

// tryRun 锁文件 防止程序多开
func tryRun() func() error {
	lock, err := lockfile.New(etc.LockFile)
	if err != nil {
		log.Fatalf("无法创建锁文件: %v", err)
	}

	err = lock.TryLock()
	if err != nil {
		if errors.Is(err, lockfile.ErrBusy) {
			log.Fatal("程序已经在运行中，请勿重复启动")
		}
		log.Fatalf("获取锁失败: %v", err)
	}

	return func() error {
		return lock.Unlock()
	}
}
