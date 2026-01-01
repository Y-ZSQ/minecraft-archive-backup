// manage/window.go
package manage

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"minecraft-archive-backup/layout/resource/icon"
	"time"
)

var App = app.New()

// WindowPool 窗口对象池
type WindowPool struct {
	pool *Pool
}

// 全局窗口池实例
var windowPool *WindowPool

// initWindowPool 初始化窗口池
func initWindowPool() {
	if windowPool == nil {
		windowPool = &WindowPool{
			pool: NewPool(
				func() interface{} {
					return newWindow()
				},
				10,            // 最大保留10个窗口
				5,             // 最小保留5个窗口
				5*time.Minute, // 每5分钟清理一次
			),
		}
	}
}

// newWindow 创建新窗口
func newWindow() fyne.Window {
	fmt.Println("创建了一个新的窗口")
	newWindow := App.NewWindow("")
	newWindow.CenterOnScreen() // 居中窗口
	newWindow.SetIcon(icon.CreeperPng)
	return newWindow
}

// GetWindow 获取窗口
func GetWindow() fyne.Window {
	initWindowPool()

	obj := windowPool.pool.Get()
	window, ok := obj.(fyne.Window)
	if !ok {
		// 如果类型断言失败，创建新窗口
		window = newWindow()
	}

	// 设置默认的窗口关闭事件
	window.SetFixedSize(true)
	window.SetCloseIntercept(func() {
		PutWindow(window)
	})

	return window
}

// PutWindow 将窗口放回池中
func PutWindow(window fyne.Window) {
	if window == nil {
		return
	}

	fmt.Println("放回了一个旧的窗口")

	// 异步执行重置操作
	fyne.Do(func() {
		// 隐藏窗口
		window.Hide()
		// 重置标题
		window.SetTitle("")
		// 移除内容
		window.SetContent(container.NewPadded())
	})

	// 放回对象池
	windowPool.pool.Put(window)
}

// GetWindowPoolSize 获取当前窗口池中的窗口数量
func GetWindowPoolSize() int {
	initWindowPool()
	return windowPool.pool.Size()
}
