package home_page

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"minecraft-archive-backup/layout/component/archive_info_page"
	"minecraft-archive-backup/layout/resource/icon"
	"minecraft-archive-backup/model/dto/database"
)

func topContent() *fyne.Container {
	// 顶部状态栏
	topTitleIcon := canvas.NewImageFromResource(icon.MinecraftPng)
	topTitleIcon.SetMinSize(fyne.NewSize(22, 22))
	topTitleIcon.FillMode = canvas.ImageFillContain

	topTitleLabel := canvas.NewText(" Minecraft 存档工具", color.White)
	topTitleLabel.TextSize = 18

	titleWithIcon := container.NewHBox(
		topTitleIcon,
		container.NewHBox(layout.NewSpacer(), layout.NewSpacer()),
		topTitleLabel,
	)

	topCreateArchiveBackupButton := widget.NewButtonWithIcon("创建存档", theme.DocumentCreateIcon(), createArchiveBackup)
	topCreateArchiveBackupButton.Importance = widget.HighImportance

	// 顶部容器
	topContainer := container.NewPadded(
		container.NewBorder(
			nil, nil,
			titleWithIcon,
			topCreateArchiveBackupButton,
		),
	)
	return topContainer
}

// createArchiveBackup 创建新的存档备份
func createArchiveBackup() {
	archive_info_page.NewWindow(&database.Archive{}, 0, refreshCard)
}
