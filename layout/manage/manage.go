package manage

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"minecraft-archive-backup/layout/resource/icon"
	"sync"
)

var App = app.New()

// Manager 窗口的创建开销是频繁的 也是巨大的
var Manager = &sync.Pool{
	New: func() interface{} {
		var newWindow = App.NewWindow("")
		fyne.Do(func() {
			fmt.Println("创建了一个新的窗口")
			newWindow.CenterOnScreen() // 居中窗口
			newWindow.SetIcon(icon.CreeperPng)
		})
		return newWindow
	},
}

func GetWindow() fyne.Window {
	var window = Manager.Get().(fyne.Window)
	// 设置默认的窗口关闭事件 如果用户直接关闭了窗口 则进行回收
	window.SetFixedSize(true) // 设置窗口大小不允许被修改
	window.SetCloseIntercept(func() {
		PutWindow(window)
	})
	return window
}

func PutWindow(window fyne.Window) {
	fmt.Println("放回了一个旧的窗口")
	// 异步执行
	fyne.Do(func() {
		// 隐藏窗口
		window.Hide()

		// 重置标题
		window.SetTitle("")

		// 移除内容
		window.SetContent(container.NewPadded())
	})

	// 放回
	Manager.Put(window)
}
