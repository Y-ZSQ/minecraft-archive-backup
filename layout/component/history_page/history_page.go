package history_page

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"minecraft-archive-backup/internal/archive"
	"minecraft-archive-backup/layout/component/progress_page"
	"minecraft-archive-backup/layout/manage"
	"minecraft-archive-backup/model/dto/database"
	"path/filepath"
	"unicode/utf8"
)

func NewWindow(a *database.Archive) {
	var window = manage.GetWindow()

	window.SetMaster()

	// 标题
	window.SetTitle("历史备份记录")

	// 内容
	window.SetContent(content(a, window))

	// 调整大小
	window.Resize(fyne.Size{Width: 430, Height: 500})

	// 展示
	window.Show()
}

func content(a *database.Archive, window fyne.Window) *fyne.Container {
	// 创建卡片容器
	grid := container.NewGridWrap(fyne.Size{Width: 400, Height: 120})

	// 添加刷新按钮
	refreshBtn := widget.NewButtonWithIcon("刷新记录", theme.ViewRefreshIcon(), func() {
		refreshCards(a, window, grid)
	})

	// 创建主容器
	mainContainer := container.NewBorder(refreshBtn, nil, nil, nil, container.NewScroll(container.NewPadded(grid)))

	// 初始加载卡片
	refreshCards(a, window, grid)

	return mainContainer
}

// 刷新卡片内容
func refreshCards(a *database.Archive, window fyne.Window, grid *fyne.Container) {
	// 清空现有卡片
	grid.RemoveAll()

	// 查询所有备份记录
	records, err := archive.GetBackupRecordsByArchiveID(a.ID)
	if err != nil {
		dialog.NewInformation("查询所有备份记录失败", err.Error(), window).Show()
		return
	}

	// 如果没有记录，显示提示信息
	if len(records) == 0 {
		grid.Add(container.NewCenter(widget.NewLabel("暂无历史备份记录")))
		return
	}

	// 遍历所有记录，为每个记录创建卡片
	for _, record := range records {
		// 格式化时间为中文格式
		formattedTime := record.CreatedAt.Format("2006年01月02日15:04:05")

		// 创建操作按钮
		deleteBtn := widget.NewButtonWithIcon("删除记录", theme.DeleteIcon(), func() {
			manage.ShowConfirmInputDialog(&manage.ConfirmInputConfig{
				Title:         "删除快照",
				Message:       fmt.Sprintf("请输入[ %s ]即可删除！", a.Name),
				ExpectedInput: a.Name,
				Parent:        window,
				Size:          fyne.Size{Width: 250, Height: 250},
				Callback: func(input string, confirmed bool) {
					if confirmed {
						// 删除 restic 存储的快照
						if err := archive.ResticForget(record.SnapShot); err != nil {
							dialog.NewInformation("删除快照失败", err.Error(), window).Show()
							return
						}

						// 删除 sqlite 中存储的快照信息
						if err := archive.DeleteBackupRecord(record.ID); err != nil {
							dialog.NewInformation("删除 sqlite 快照失败", err.Error(), window).Show()
							return
						}

						// 删除后刷新卡片
						refreshCards(a, window, grid)
					}
				},
			})
		})
		deleteBtn.Importance = widget.DangerImportance

		restoreBtn := widget.NewButtonWithIcon("快照回档", theme.ViewRefreshIcon(), func() {
			fmt.Println(filepath.Join(a.Path, "session.lock"))
			if ok, _ := IsWorldInUse(filepath.Join(a.Path, "session.lock")); ok {
				dialog.NewInformation("注意", "如果您要回档,您需要退出当前地图(无需退出游戏)", window).Show()
				return
			}

			manage.ShowConfirmInputDialog(&manage.ConfirmInputConfig{
				Title:         "回档指定快照",
				Message:       "您确认要进行回档吗？",
				ExpectedInput: "确认回档",
				Placeholder:   "请输入确认回档",
				ErrorTest:     "请注意！回档的操作是不可逆的，务必谨慎！",
				Parent:        window,
				Size:          fyne.Size{Width: 250, Height: 250},
				Callback: func(input string, confirmed bool) {
					if confirmed {
						// 进行回档
						var stdChan = archive.ResticRestore(a, &record)
						// 将通道和存档信息传入备份页面
						progress_page.NewWindow(a, 1, stdChan, func(success bool, errorMsg string, lastMessage *archive.BackupMessage) {
							if !success {
								dialog.NewInformation("回档失败", err.Error(), window).Show()
								return
							} else {
								dialog.NewInformation("回档成功", "回档成功啦，感谢您的使用！", window).Show()
							}
						})
					}
					//		})
					//	}
				},
			})
		})
		restoreBtn.Importance = widget.HighImportance

		// 详细信息按钮
		infoBtn := widget.NewButtonWithIcon("存档信息", theme.ListIcon(), func() {
			var rawData, _ = archive.ResticRawData(record.SnapShot)
			var RestoreSize, _ = archive.ResticRestoreSize(record.SnapShot)
			dialog.NewInformation("存档详细信息",
				fmt.Sprintf(`存档备注：%s
创建时间：%s
快照ID ：%s
存储占用：%dMB
原始大小：%dMB
`, record.Comment, record.CreatedAt.Format("2006年01月02日15:04:05"), record.SnapShot[:8], rawData.TotalSize/1048576, RestoreSize.TotalSize/1048576),
				window,
			).Show()
		})
		infoBtn.Importance = widget.WarningImportance

		// 创建卡片
		card := widget.NewCard(
			formattedTime,
			truncateWithEllipsis(record.Comment, 27),
			container.NewHBox(deleteBtn, infoBtn, restoreBtn),
		)

		// 将卡片添加到网格中
		grid.Add(card)
	}

	// 刷新容器显示
	fyne.Do(func() {
		grid.Refresh()
	})
}

func truncateWithEllipsis(s string, maxChars int) string {
	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}

	runes := []rune(s)
	return string(runes[:maxChars-1]) + "…" // 使用省略号
}
