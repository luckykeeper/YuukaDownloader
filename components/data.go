// Based on fyne_demo, modified for yuukaDownloader
// 此文件定义了菜单组件（左侧功能切换组件）的信息
package components

import (
	"fyne.io/fyne/v2"
)

// Component defines the data structure
type Component struct {
	Title, Intro string
	View         func(w fyne.Window) fyne.CanvasObject
	SupportWeb   bool
}

var (
	// Components defines the metadata for each component
	Components = map[string]Component{
		"welcome": {"Welcome yuukaDown!", "雷猴哇~<(￣︶￣)↗[GO!]", welcomeScreen, true},
		"下载平台设定": {"下载平台设定",
			"在这里设定下载平台参数 ！ヾ(≧▽≦*)o —— ",
			downloaderPlatformSettings,
			true,
		},
		"下发下载任务": {"下发下载任务",
			"下发下载（下崽）任务到平台^_~",
			downloadQuests,
			true,
		},
		"Aria2": {"Aria2",
			"在这里设定下载平台参数 ！ヾ(≧▽≦*)o —— Aria2 | 设置完成后记得保存和测试~",
			Aria2PlatformSetting,
			true,
		},
		"历史数据库": {"历史数据库",
			"过往的下载任务可以在这里看到",
			historyDatabase,
			true,
		},
	}

	// ComponentIndex  defines how our Components should be laid out in the index tree
	ComponentIndex = map[string][]string{
		"":       {"welcome", "下载平台设定", "下发下载任务", "历史数据库"},
		"下载平台设定": {"Aria2"},
	}
)
