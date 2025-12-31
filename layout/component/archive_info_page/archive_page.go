package archive_info_page

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"minecraft-archive-backup/internal/archive"
	"minecraft-archive-backup/layout/manage"
	"minecraft-archive-backup/model/dto/database"
)

// OperationMode 操作模式
type OperationMode int

const (
	ModeCreate OperationMode = iota // 创建模式
	ModeEdit                        // 编辑模式
)

func NewWindow(info *database.Archive, mode OperationMode, refreshCallback func()) {
	window := manage.GetWindow()

	// 调整大小
	window.Resize(fyne.NewSize(420, 445))

	// 设置标题
	var title string
	switch mode {
	case ModeCreate:
		title = "创建存档"
	case ModeEdit:
		title = "编辑存档信息"
	}
	window.SetTitle(title)

	// 内容
	window.SetContent(newContent(window, info, mode, refreshCallback))

	// 展示
	window.Show()
}

func newContent(window fyne.Window, info *database.Archive, mode OperationMode, refreshCallback func()) fyne.CanvasObject {
	// 顶部标题
	var titleText string
	switch mode {
	case ModeCreate:
		titleText = "创建存档"
	case ModeEdit:
		titleText = fmt.Sprintf("编辑 [ %s ] 信息", info.Name)
	}

	title := widget.NewLabelWithStyle(titleText, fyne.TextAlignCenter, fyne.TextStyle{
		Bold: true,
	})

	// 存档名称
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("输入存档名称")
	nameEntry.Text = info.Name
	nameEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("存档名称不能为空")
		}
		return nil
	}

	// 存档注释
	commentEntry := widget.NewMultiLineEntry()
	commentEntry.SetPlaceHolder("输入存档注释（可选）")
	commentEntry.Wrapping = fyne.TextWrapWord
	commentEntry.SetMinRowsVisible(4)
	commentEntry.Text = info.Comment

	// 存档路径
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("输入存档文件夹路径")
	pathEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("存档路径不可为空")
		}
		return nil
	}
	pathEntry.Text = info.Path

	// 取消按钮
	cancelButton := widget.NewButtonWithIcon("取消", theme.CancelIcon(), func() {
		manage.PutWindow(window)
	})

	// 保存按钮
	var saveButtonText string
	if mode == ModeCreate {
		saveButtonText = "创建"
	} else {
		saveButtonText = "保存"
	}

	saveButton := widget.NewButtonWithIcon(saveButtonText, theme.ConfirmIcon(), func() {
		// 验证输入
		if nameEntry.Text == "" {
			dialog.NewInformation("注意！", "存档名称为空", window).Show()
			return
		}

		if pathEntry.Text == "" {
			dialog.NewInformation("注意！", "存档路径为空", window).Show()
			return
		}

		if !IsValidPathFormat(pathEntry.Text) {
			dialog.NewInformation("注意！", "存档路径格式有误", window).Show()
			return
		}

		// 更新info对象
		info.Name = nameEntry.Text
		info.Comment = commentEntry.Text
		info.Path = pathEntry.Text

		// 根据模式执行不同操作
		var err error
		if mode == ModeCreate {
			err = createArchive(info)
		} else {
			// 编辑模式，调用编辑函数
			err = editArchive(info)
		}

		if err != nil {
			// 根据模式显示不同的错误信息
			if mode == ModeCreate {
				dialog.NewInformation("创建失败", err.Error(), window).Show()
			} else {
				dialog.NewInformation("保存失败", err.Error(), window).Show()
			}
			return
		}

		// 执行回调刷新函数
		if refreshCallback != nil {
			refreshCallback()
		}

		// 将窗口放回对象池
		manage.PutWindow(window)
	})
	saveButton.Importance = widget.HighImportance

	// 按钮容器
	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelButton,
		layout.NewSpacer(),
		saveButton,
		layout.NewSpacer(),
	)

	// 创建主容器
	mainContainer := container.NewVBox(
		// 标题
		container.NewPadded(title),

		// 存档名称
		container.NewVBox(
			widget.NewLabel("存档名称"),
			layout.NewSpacer(),
			nameEntry,
		),

		// 注释
		container.NewVBox(
			widget.NewLabel("注释"),
			layout.NewSpacer(),
			commentEntry,
		),

		// 存档路径
		container.NewVBox(
			widget.NewLabel("存档路径"),
			layout.NewSpacer(),
			pathEntry,
		),

		layout.NewSpacer(),
		container.NewPadded(buttonContainer),
	)

	return container.NewPadded(mainContainer)
}

// 创建存档
func createArchive(info *database.Archive) error {
	newInfo, err := archive.GetOrCreateArchiveCache(0, func() (*database.Archive, error) {
		err := archive.CreateArchive(info)
		return info, err
	})
	info = newInfo
	return err
}

// 编辑存档
func editArchive(info *database.Archive) error {
	err := archive.UpdateArchive(info)
	if err == nil {
		_ = archive.StoreArchiveCache(info)
	}
	fmt.Println("更新存档", info)
	fmt.Println(archive.LoadArchiveCache(info.ID))
	return err
}
