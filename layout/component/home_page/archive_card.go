package home_page

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"minecraft-archive-backup/internal/archive"
	"minecraft-archive-backup/layout/component/archive_info_page"
	"minecraft-archive-backup/layout/component/history_page"
	"minecraft-archive-backup/layout/component/progress_page"
	"minecraft-archive-backup/layout/manage"
	"minecraft-archive-backup/model/dto/database"
	"strings"

	"unicode/utf8"
)

type ArchiveCard struct {
	content     *widget.Card
	archiveInfo *database.Archive
}

func NewCard(a *database.Archive, window fyne.Window) *ArchiveCard {
	// 删除存档按钮
	var deleteBtn = widget.NewButtonWithIcon("删除存档", theme.DeleteIcon(), func() {
		manage.ShowConfirmInputDialog(&manage.ConfirmInputConfig{
			Title:         "删除存档",
			Message:       fmt.Sprintf("请输入[ %s ]即可删除！", a.Name),
			ExpectedInput: a.Name,
			Parent:        window,
			Size:          fyne.Size{Width: 250, Height: 250},
			Callback: func(input string, confirmed bool) {
				// 输入正确 执行删除操作
				if confirmed {
					// 删除 restic 中相关的快照
					if records, err := archive.GetBackupRecordsByArchiveID(a.ID); err != nil {
						dialog.NewInformation(fmt.Sprintf("获取[ %s ]备份记录失败", a.Name), err.Error(), window).Show()
						return
					} else {
						var snapShots = make([]string, len(records))
						for _, r := range records {
							snapShots = append(snapShots, r.SnapShot)
						}
						if err := archive.ResticForget(snapShots...); err != nil {
							dialog.NewInformation("删除 restic 快照失败", err.Error(), window).Show()
							return
						}
					}
					// 删除 sqlite 中存储的数据
					var err = archive.DeleteArchive(a.ID)
					if err != nil {
						fmt.Println(err)
						dialog.NewInformation("删除 sqlite 失败", err.Error(), window).Show()
						return
					}
					// 删除缓存
					_ = archive.DeleteArchiveCache(a.ID)

					// 刷新卡片
					refreshCard()
				}
			},
		})
	})
	deleteBtn.Importance = widget.DangerImportance

	// 设置信息按钮
	var editBtn = widget.NewButtonWithIcon("设置信息", theme.ListIcon(), func() {
		archive_info_page.NewWindow(a, 1, refreshCard)
	})
	editBtn.Importance = widget.WarningImportance

	// 历史记录按钮
	var historyBtn = widget.NewButtonWithIcon("历史记录", theme.HistoryIcon(), func() {
		history_page.NewWindow(a)
	})
	historyBtn.Importance = widget.SuccessImportance

	// 存档备份按钮
	var backupBtn = widget.NewButtonWithIcon("存档备份", theme.DocumentSaveIcon(), func() {
		// 创建输入对话框
		entry := widget.NewEntry()
		entry.SetPlaceHolder("可选")

		var dlg = dialog.NewForm("自定义存档备份，有助于您更好的寻找您的存档历史!",
			"开始备份",
			"取消",
			[]*widget.FormItem{
				widget.NewFormItem("备注：", entry),
			},
			func(confirm bool) {
				if !confirm {
					return
				}
				// 获取用户输入的备注
				comment := strings.TrimSpace(entry.Text)

				// 执行备份
				var stdChan = archive.ResticBackup(a)

				// 将通道和存档信息传入备份页面
				progress_page.NewWindow(a, 0, stdChan, func(success bool, errorMsg string, lastMessage *archive.BackupMessage) {
					if success {
						// 创建一个备份记录
						var err = archive.CreateBackupRecord(&database.BackupRecord{
							ArchiveID: a.ID,
							SnapShot:  lastMessage.SnapshotID,
							Comment:   comment,
						})
						if err != nil {
							dialog.NewInformation("快照ID写入到sqlite失败", err.Error(), window).Show()
							return
						}
					}
				})
			},
			window,
		)
		dlg.Resize(fyne.NewSize(350, 150))
		dlg.Show()
	})
	backupBtn.Importance = widget.HighImportance

	return &ArchiveCard{
		archiveInfo: a,
		content: widget.NewCard(truncateWithEllipsis(a.Name, 15), truncateWithEllipsis(a.Comment, 27),
			container.NewHBox(
				deleteBtn,
				editBtn,
				historyBtn,
				backupBtn,
			)),
	}
}

func truncateWithEllipsis(s string, maxChars int) string {
	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}

	runes := []rune(s)
	return string(runes[:maxChars-1]) + "…" // 使用省略号
}
