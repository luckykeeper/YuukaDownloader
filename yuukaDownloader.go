// 初音站（yngal/fufugal）链接快速生成，并送 Aria2 下载
// 适用于多part，多站链接快速生成复制
// Powered By Luckykeeper 20230411 ver 1.0.0, Written By Go 1.20.3

package main

import (
	"database/sql"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"

	"yuukaDownloader/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"

	_ "github.com/mattn/go-sqlite3"
)

const (
	sql_initialize_yuukaDown     = `CREATE TABLE yuukaDown (yuukaFunction TEXT PRIMARY KEY NOT NULL,yuukaStatus BOOLEAN NOT NULL);`
	sql_initialize_yuukaDownInit = `INSERT INTO yuukaDown (yuukaFunction,yuukaStatus) VALUES ('yuukaDownHistory', true);`
	sql_initialize_platform      = `CREATE TABLE yuukaDownPlatform (downloaderPlatform TEXT PRIMARY KEY NOT NULL,platformUrl TEXT,jsonRPCVer TEXT,ariaToken TEXT);`
	sql_initialize_platformInit  = `INSERT INTO yuukaDownPlatform (downloaderPlatform) VALUES ('Aria2');`
	sql_initialize_galDB         = `CREATE TABLE yuukaDownGalDB (galgameName TEXT PRIMARY KEY NOT NULL,downloadBaseUrl TEXT NOT NULL,partNum TEXT NOT NULL,fileType TEXT NOT NULL,subArea TEXT NOT NULL);`

	preferenceCurrentComponent = "currentComponent"
)

var topWindow fyne.Window

// 显示中文
// 设置环境变量   通过 go-findfont 寻找simkai.ttf 字体
func init() {
	// Windows Platform 使用系统变量内的字体
	if runtime.GOOS == "windows" {
		log.Println("Windows Platform, Import System Font, Program Init...")
		fontPaths := findfont.List()
		for _, path := range fontPaths {
			if strings.Contains(path, "simkai.ttf") {
				os.Setenv("FYNE_FONT", path) // 设置环境变量
				break
			}
		}
		// Linux Platform 使用程序提供的字体，Ren'Py 同款
	} else if runtime.GOOS == "linux" {
		log.Println("Linux Platform, Use Program Font, Program Init...")
		os.Setenv("FYNE_FONT", "./SourceHanSansLite.ttf") // 设置环境变量
	} else {
		// 其它平台，使用程序提供的字体，Ren'Py 同款
		log.Println("Other Platform, Use Program Font, Program Init...")
		os.Setenv("FYNE_FONT", "./SourceHanSansLite.ttf") // 设置环境变量
	}
	// 初始化数据库
	if exists, _ := PathExists("./yuukaDown.db"); exists {
		log.Println("检测数据库存在，跳过数据库初始化任务")
	} else {
		log.Println("开始任务 -> 创建并初始化数据库！")
		db, _ := sql.Open("sqlite3", "./yuukaDown.db")
		defer db.Close()
		db.Exec(sql_initialize_yuukaDown)     // 创建 yuukaDown 自身数据库
		db.Exec(sql_initialize_yuukaDownInit) // 初始化 yuukaDown 自身数据库
		db.Exec(sql_initialize_platform)      // 创建下载平台数据库
		db.Exec(sql_initialize_platformInit)  // 初始化下载平台数据库
		db.Exec(sql_initialize_galDB)         // 初始化 Galgame 数据库
		log.Println("创建并初始化数据库完成！")
	}
	log.Println("Init done! Program starting...")
}

func main() {
	a := app.NewWithID("yuukadown.luckykeeper.site")
	logo, _ := fyne.LoadResourceFromPath("yuukaIcon.ico")
	a.SetIcon(logo)
	makeTray(a)
	logLifecycle(a)
	w := a.NewWindow("yuukaDown, A software to Download Gal Through muti-Platform | Powered by Luckykeeper | Build 20230420 | Ver 1.0.0")
	topWindow = w
	w.SetMainMenu(makeMenu(a, w))
	w.SetMaster()

	content := container.NewMax()
	title := widget.NewLabel("Component name")
	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
	intro.Wrapping = fyne.TextWrapWord
	setComponent := func(t components.Component) {
		if fyne.CurrentDevice().IsMobile() {
			child := a.NewWindow(t.Title)
			topWindow = child
			child.SetContent(t.View(topWindow))
			child.Show()
			child.SetOnClosed(func() {
				topWindow = w
			})
			return
		}

		title.SetText(t.Title)
		intro.SetText(t.Intro)

		content.Objects = []fyne.CanvasObject{t.View(w)}
		content.Refresh()
	}

	component := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), intro), nil, nil, nil, content)
	if fyne.CurrentDevice().IsMobile() {
		w.SetContent(makeNav(setComponent, false))
	} else {
		split := container.NewHSplit(makeNav(setComponent, true), component)
		split.Offset = 0.2
		w.SetContent(split)
	}
	w.Resize(fyne.NewSize(1280, 720))
	w.ShowAndRun()
}

// 生命周期日志
func logLifecycle(a fyne.App) {
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	a.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

// 顶部菜单
func makeMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	aboutMenu := fyne.NewMenu("关于",
		fyne.NewMenuItem("访问作者博客", func() {
			u, _ := url.Parse("https://luckykeeper.site")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("访问 Github —— 查看我的其它开源项目", func() {
			u, _ := url.Parse("https://github.com/luckykeeper")
			_ = a.OpenURL(u)
		}),
	)

	main := fyne.NewMainMenu(
		aboutMenu,
	)
	return main
}

// 任务栏托盘
func makeTray(a fyne.App) {
	if desk, ok := a.(desktop.App); ok {
		h := fyne.NewMenuItem("yuukaDown, A software to Download Gal Through muti-Platform", func() {})
		menu := fyne.NewMenu("yuukaDown By Luckykeeper", h)
		h.Action = func() {
			log.Println("Tray Menu Clicked!")
			h.Label = "yuukaDown, A software to Download Gal Through muti-Platform"
			u, _ := url.Parse("https://github.com/luckykeeper")
			a.OpenURL(u)
			menu.Refresh()
		}
		desk.SetSystemTrayMenu(menu)
	}
}

func unsupportedComponent(t components.Component) bool {
	return !t.SupportWeb && fyne.CurrentDevice().IsBrowser()
}

func makeNav(setComponent func(component components.Component), loadPrevious bool) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return components.ComponentIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := components.ComponentIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := components.Components[uid]
			if !ok {
				fyne.LogError("Missing component panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
			if unsupportedComponent(t) {
				obj.(*widget.Label).TextStyle = fyne.TextStyle{Italic: true}
			} else {
				obj.(*widget.Label).TextStyle = fyne.TextStyle{}
			}
		},
		OnSelected: func(uid string) {
			if t, ok := components.Components[uid]; ok {
				if unsupportedComponent(t) {
					return
				}
				a.Preferences().SetString(preferenceCurrentComponent, uid)
				setComponent(t)
			}
		},
	}

	if loadPrevious {
		currentPref := a.Preferences().StringWithFallback(preferenceCurrentComponent, "welcome")
		tree.Select(currentPref)
	}

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}

// 判断文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
