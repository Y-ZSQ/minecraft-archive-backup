package home_page

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"minecraft-archive-backup/layout/manage"
	"os"
)

func Run() {
	var window = manage.GetWindow()

	// 赋值 refreshCard
	refreshCard = refreshCardHelp(window)

	// 首页的 Window 为主窗口
	window.SetMaster()

	// 标题
	window.SetTitle("Minecraft Archive Backup Tool")

	// 内容
	window.SetContent(content())

	// 调整大小
	window.Resize(fyne.Size{Width: 430, Height: 500})

	// 关闭窗口时 直接停止程序
	window.SetCloseIntercept(func() {
		os.Exit(0)
	})

	// 按键监听
	window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeyF5:
			refreshCard()
			dialog.NewInformation("refresh", "刷新完毕", window).Show()
		}
	})

	// 展示所有的卡片
	refreshCard()

	// 展示并且执行 并会阻塞代码
	window.ShowAndRun()
}

func content() *fyne.Container {
	return container.NewBorder(topContent(), nil, nil, nil, container.NewScroll(container.NewPadded(girdContainer)))
}
