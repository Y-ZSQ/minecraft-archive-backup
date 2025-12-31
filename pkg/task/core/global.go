package core

import (
	"log"
	"time"
)

var (
	// TaskMenger 管理所有的定时任务
	TaskMenger = NewTaskMachine()
)

// init 初始化 TaskMenger 并且 启动一个携程 循环的执行所有的定时任务
func init() {
	// 添加 RemoveExpiredImage 删除 过期的图片 任务
	//TaskMenger.AddTask(&remove_expired_image.RemoveExpiredImage{})

	// 在 Range 内部启动 携程 循环的执行定时任务
	TaskMenger.Range(func(task Task) {
		// 判断是否立即执行
		if task.ExecuteImmediately() {
			log.Printf("[ %s ]执行任务!\n", task.Key())
			task.Run()
		}
		// 定时器
		var tick = time.NewTicker(task.Interval())
		for {
			select {
			case <-tick.C:
				log.Printf("[ %s ]执行任务!\n", task.Key())
				task.Run()
			}
		}
	})
}
