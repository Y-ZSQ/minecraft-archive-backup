package home_page

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"minecraft-archive-backup/internal/archive"
	"sort"
)

var girdContainer = container.NewGridWrap(fyne.Size{Width: 400, Height: 120})

// refreshCard 在 home 页面 会进行赋值 并传入 window 参数
var refreshCard func()

func refreshCardHelp(window fyne.Window) func() {
	return func() {
		girdContainer.RemoveAll()

		// 获取所有的应用
		var archives = archive.LoadAllArchiveCache()

		// 提取所有的键
		keys := make([]uint, 0, len(archives))
		for k := range archives {
			keys = append(keys, k)
		}

		// 对键进行排序
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j] // 升序
			// 降序: return keys[i] > keys[j]
		})

		// 按排序后的键顺序添加
		for _, key := range keys {
			a := archives[key]
			girdContainer.Add(NewCard(a, window).content)
		}

		girdContainer.Refresh()
	}
}
