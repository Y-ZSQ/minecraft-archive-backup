package manage

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ConfirmInputConfig 确认输入对话框配置
type ConfirmInputConfig struct {
	Title         string          // 对话框标题
	Message       string          // 提示信息
	ExpectedInput string          // 预期用户输入的内容
	ConfirmText   string          // 确认按钮文字，默认为"确认"
	CancelText    string          // 取消按钮文字，默认为"取消"
	Callback      ConfirmCallback // 回调函数
	Parent        fyne.Window     // 父窗口
	ErrorTest     string          // 错误标签文字,默认为"一旦删除将无法再次复原！"
	Placeholder   string          // 输入框占位符，可选
	Size          fyne.Size       // 对话框大小，可选
}

// ConfirmCallback 回调函数类型
type ConfirmCallback func(input string, confirmed bool)

// ShowConfirmInputDialog 显示确认输入对话框
func ShowConfirmInputDialog(config *ConfirmInputConfig) {
	// 设置默认值
	if config.ConfirmText == "" {
		config.ConfirmText = "确认"
	}
	if config.CancelText == "" {
		config.CancelText = "取消"
	}
	if config.Placeholder == "" {
		config.Placeholder = "请输入: " + config.ExpectedInput
	}
	if config.Size.IsZero() {
		config.Size = fyne.NewSize(400, 200)
	}
	if config.ErrorTest == "" {
		config.ErrorTest = "一旦删除将无法再次复原！"
	}

	// 输入框
	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder(config.Placeholder)

	// 错误提示标签
	errorLabel := widget.NewLabel(config.ErrorTest)
	errorLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建自定义对话框
	var dlg dialog.Dialog

	// 确认按钮点击处理
	onConfirm := func() {
		userInput := inputEntry.Text

		// 调用回调函数
		var verify = userInput == config.ExpectedInput
		if config.Callback != nil {
			config.Callback(userInput, verify)
		}

		// 验证输入
		if !verify {
			errorLabel.SetText("输入不正确，请重新输入")
			errorLabel.Show()
			// 不关闭对话框，只刷新错误提示
			dlg.Refresh()
			return
		}

		// 输入正确，关闭对话框并执行回调
		dlg.Hide()
	}

	// 取消按钮点击处理
	onCancel := func() {
		dlg.Hide()
		if config.Callback != nil {
			config.Callback("", false)
		}
	}

	// 创建按钮
	confirmBtn := widget.NewButton(config.ConfirmText, onConfirm)
	confirmBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton(config.CancelText, onCancel)

	// 按钮容器
	buttonContainer := container.NewBorder(
		nil, nil,
		cancelBtn,
		confirmBtn,
	)

	// 表单内容
	content := container.NewVBox(
		widget.NewLabel(config.Message),
		widget.NewSeparator(),
		inputEntry,
		errorLabel,
		widget.NewSeparator(),
		buttonContainer,
	)

	// 创建自定义对话框
	dlg = dialog.NewCustomWithoutButtons(config.Title, content, config.Parent)

	// 设置对话框大小
	dlg.Resize(config.Size)

	// 添加Enter键支持
	inputEntry.OnSubmitted = func(s string) {
		onConfirm()
	}

	// 显示对话框
	dlg.Show()
}
