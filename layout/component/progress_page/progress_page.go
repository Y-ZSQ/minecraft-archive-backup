package progress_page

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"minecraft-archive-backup/internal/archive"
	"minecraft-archive-backup/layout/manage"
	"minecraft-archive-backup/model/dto/database"
	"strings"
	"time"
)

// Mode 定义了进度页面的模式
type Mode int

const (
	ModeBackup  Mode = iota // 备份模式
	ModeRestore Mode = 1    // 回档模式
)

// CompletionCallback 回调函数类型
// 参数: 1. 是否成功 2. 错误信息(如果失败) 3. 最后一条消息(可能包含快照ID等信息)
type CompletionCallback func(success bool, errorMsg string, lastMessage *archive.BackupMessage)

// 获取模式对应的标题
func getTitleByMode(mode Mode) string {
	switch mode {
	case ModeBackup:
		return "正在备份中"
	case ModeRestore:
		return "正在回档中"
	}
	return ""
}

// ProgressWindow 进度窗口结构体
type ProgressWindow struct {
	window      fyne.Window
	mode        Mode
	progress    *widget.ProgressBar
	statusText  *widget.Label
	details     *widget.Label
	dotLabel    *widget.Label
	done        chan bool
	callback    CompletionCallback     // 回调函数
	completed   bool                   // 标记是否已完成
	resultSent  bool                   // 标记结果是否已发送
	lastMessage *archive.BackupMessage // 记录最后一条消息
	archive     *database.Archive      // 存档信息
}

// NewWindow 创建进度窗口
// callback: 回调函数，在完成时调用
func NewWindow(a *database.Archive, mode Mode, state <-chan *archive.BackupMessage, callback CompletionCallback) {
	window := manage.GetWindow()

	// 调整窗口大小
	window.Resize(fyne.NewSize(400, 400))

	// 设置窗口标题
	title := getTitleByMode(mode)
	window.SetTitle(title)

	// 创建进度窗口实例
	pw := &ProgressWindow{
		window:      window,
		mode:        mode,
		done:        make(chan bool),
		callback:    callback,
		completed:   false,
		resultSent:  false,
		lastMessage: nil,
		archive:     a, // 保存存档信息
	}

	// 初始化UI
	pw.initUI()

	// 设置窗口内容
	fyne.Do(func() {
		window.SetContent(pw.content())
	})

	// 启动点动画
	go pw.startDotAnimation()

	// 启动状态监听
	go pw.listenState(state)

	// 监听窗口关闭事件
	window.SetCloseIntercept(func() {
		if !pw.resultSent && pw.callback != nil {
			// 如果窗口被手动关闭，通知外部失败
			pw.callback(false, "用户取消了操作", pw.lastMessage)
		}
		go func() {
			time.Sleep(time.Millisecond * 1500)
			manage.PutWindow(window)
		}()

	})

	// 展示窗口
	fyne.Do(func() {
		window.Show()
	})
}

// initUI 初始化UI组件
func (pw *ProgressWindow) initUI() {
	// 创建进度条
	pw.progress = widget.NewProgressBar()
	pw.progress.Min = 0
	pw.progress.Max = 100
	pw.progress.TextFormatter = func() string {
		return fmt.Sprintf("%.1f%%", pw.progress.Value*100)
	}

	// 创建状态文本标签
	pw.statusText = widget.NewLabel("等待开始...")
	pw.statusText.TextStyle = fyne.TextStyle{Bold: true}

	// 创建详细信息标签
	pw.details = widget.NewLabel("")
	pw.details.Wrapping = fyne.TextWrapWord

	// 创建动态点标签
	pw.dotLabel = widget.NewLabel("")
}

// content 构建UI内容
func (pw *ProgressWindow) content() *fyne.Container {
	// 创建顶部标题区域
	titleLabel := widget.NewLabel(getTitleByMode(pw.mode))
	titleLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	// 创建动态点容器
	dotContainer := container.NewHBox(
		widget.NewLabel("处理中"),
		pw.dotLabel,
	)

	// 构建完整的UI布局
	return container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		pw.statusText,
		container.NewVBox(
			widget.NewLabel("进度:"),
			pw.progress,
		),
		dotContainer,
		widget.NewSeparator(),
		widget.NewLabel("详细信息:"),
		pw.details,
	)
}

// startDotAnimation 启动动态点动画
func (pw *ProgressWindow) startDotAnimation() {
	dots := []string{"", ".", "..", "...", ".."}
	index := 0

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-pw.done:
			return
		case <-ticker.C:
			fyne.Do(func() {
				pw.dotLabel.SetText(dots[index])
			})
			index = (index + 1) % len(dots)
		}
	}
}

// listenState 监听备份/回档状态
func (pw *ProgressWindow) listenState(state <-chan *archive.BackupMessage) {
	defer func() {
		pw.done <- true
		close(pw.done)
	}()

	for msg := range state {
		pw.updateUI(msg)

		// 记录最后一条消息
		pw.lastMessage = msg

		// 检查是否完成
		if msg.MessageType == "summary" || msg.MessageType == "done" {
			// 延迟关闭窗口
			pw.complete(true, "", msg) // 成功完成，传递最后的消息
			return
		}

		// 检查是否有错误
		if msg.Code != 0 || (msg.MessageType == "error" && msg.Message != "") {
			pw.complete(false, msg.Message, msg) // 失败，传递错误消息
			pw.showError(msg)
			return
		}
	}

	// 如果通道关闭但没有收到完成消息
	pw.complete(false, "操作被意外终止", pw.lastMessage)
}

// updateUI 根据消息更新UI
func (pw *ProgressWindow) updateUI(msg *archive.BackupMessage) {
	fyne.Do(func() {
		// 更新进度条
		if msg.PercentDone > 0 {
			pw.progress.SetValue(msg.PercentDone)
		}

		// 更新状态文本
		var statusParts []string

		switch msg.MessageType {
		case "status":
			statusParts = append(statusParts, "扫描文件中...")
			if msg.TotalFiles > 0 {
				statusParts = append(statusParts,
					fmt.Sprintf("文件: %d/%d", msg.FilesDone, msg.TotalFiles))
			}
			if msg.TotalBytes > 0 {
				mbDone := float64(msg.BytesDone) / 1024 / 1024
				mbTotal := float64(msg.TotalBytes) / 1024 / 1024
				statusParts = append(statusParts,
					fmt.Sprintf("数据: %.1f/%.1f MB", mbDone, mbTotal))
			}

		case "summary":
			statusParts = append(statusParts, "正在创建快照...")
			if msg.TotalDuration > 0 {
				statusParts = append(statusParts,
					fmt.Sprintf("耗时: %.1fs", msg.TotalDuration))
			}

		case "done":
			statusParts = append(statusParts, "完成!")

		default:
			statusParts = append(statusParts, "处理中...")
		}
		pw.statusText.SetText(strings.Join(statusParts, " | "))

		// 更新详细信息
		var details []string

		// 文件统计
		if msg.FilesNew > 0 || msg.FilesChanged > 0 || msg.FilesUnmodified > 0 {
			details = append(details,
				fmt.Sprintf("文件: 新增(%d) 修改(%d) 未变(%d)",
					msg.FilesNew, msg.FilesChanged, msg.FilesUnmodified))
		}

		// 目录统计
		if msg.DirsNew > 0 || msg.DirsChanged > 0 || msg.DirsUnmodified > 0 {
			details = append(details,
				fmt.Sprintf("目录: 新增(%d) 修改(%d) 未变(%d)",
					msg.DirsNew, msg.DirsChanged, msg.DirsUnmodified))
		}

		// 数据统计
		if msg.DataAdded > 0 {
			mbAdded := float64(msg.DataAdded) / 1024 / 1024
			details = append(details, fmt.Sprintf("新增数据: %.2f MB", mbAdded))
		}

		// 进度信息
		if msg.PercentDone > 0 {
			details = append(details, fmt.Sprintf("进度: %.1f%%", msg.PercentDone*100))
		}

		// 如果有快照ID，显示
		if msg.SnapshotID != "" {
			details = append(details, fmt.Sprintf("快照ID: %s", msg.SnapshotID))
		}

		// 如果有消息内容，显示
		if msg.Message != "" && msg.MessageType != "status" {
			details = append(details, fmt.Sprintf("消息: %s", msg.Message))
		}

		if len(details) > 0 {
			pw.details.SetText(strings.Join(details, "\n"))
		}
	})
}

// showError 显示错误信息
func (pw *ProgressWindow) showError(msg *archive.BackupMessage) {
	fyne.Do(func() {
		pw.statusText.SetText("错误!")
		pw.statusText.TextStyle = fyne.TextStyle{Bold: true}

		errorMsg := "发生错误"
		if msg.Message != "" {
			errorMsg = msg.Message
		}

		pw.details.SetText(fmt.Sprintf("错误代码: %d\n%s", msg.Code, errorMsg))
	})
}

// complete 完成处理
func (pw *ProgressWindow) complete(success bool, errorMsg string, lastMsg *archive.BackupMessage) {
	// 防止重复调用
	if pw.completed {
		return
	}
	pw.completed = true

	// 如果传入了最后消息，更新
	if lastMsg != nil {
		pw.lastMessage = lastMsg
	}

	fyne.Do(func() {
		if success {
			pw.statusText.SetText("完成!")
			pw.progress.SetValue(1.0)
			pw.dotLabel.SetText("✓")

			// 显示完成详情
			var details []string
			details = append(details, "操作成功完成")

			if pw.lastMessage != nil {
				// 显示快照信息
				if pw.lastMessage.SnapshotID != "" {
					details = append(details, fmt.Sprintf("快照ID: %s", pw.lastMessage.SnapshotID))
				}
				if pw.lastMessage.Message != "" {
					details = append(details, fmt.Sprintf("消息: %s", pw.lastMessage.Message))
				}
			}

			pw.details.SetText(strings.Join(details, "\n"))
		} else {
			pw.statusText.SetText("失败!")
			pw.statusText.TextStyle = fyne.TextStyle{Bold: true}
			pw.dotLabel.SetText("✗")
			if errorMsg != "" {
				pw.details.SetText("错误: " + errorMsg)
			}
		}
	})

	// 调用回调函数
	pw.notifyResult(success, errorMsg, pw.lastMessage)

	go func() {
		time.Sleep(time.Millisecond * 1500)
		manage.PutWindow(pw.window)
	}()
}

// notifyResult 通知结果
func (pw *ProgressWindow) notifyResult(success bool, errorMsg string, lastMsg *archive.BackupMessage) {
	if pw.resultSent || pw.callback == nil {
		return
	}
	pw.resultSent = true

	// 在单独的goroutine中调用回调，避免阻塞UI
	go func() {
		// 等待一小段时间确保UI已更新
		time.Sleep(100 * time.Millisecond)
		pw.callback(success, errorMsg, lastMsg)
	}()
}
