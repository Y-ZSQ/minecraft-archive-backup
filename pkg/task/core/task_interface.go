package core

import (
	"sync"
	"time"
)

// TaskMachine 统一管理所有的任务 并按照间隔 循环执行(执行任务时 使用携程)
type TaskMachine struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

type Task interface {
	// Key 当前任务机器的名称
	Key() string
	// Interval 执行间隔
	Interval() time.Duration
	// Run 运行一次该任务
	Run()
	// ExecuteImmediately 是否在程序刚启动时 就立即执行
	ExecuteImmediately() bool
}

func NewTaskMachine() *TaskMachine {
	return &TaskMachine{
		tasks: make(map[string]Task),
	}
}

// AddTask 新增一个任务
func (t *TaskMachine) AddTask(newTask Task) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tasks[newTask.Key()] = newTask
}

// RemoveTask 移除一个任务
func (t *TaskMachine) RemoveTask(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tasks, key)
}

// Range 遍历所有任务，并对每个任务执行 fn（线程安全）
func (t *TaskMachine) Range(fn func(task Task)) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, task := range t.tasks {
		go fn(task)
	}
}
