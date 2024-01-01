// 下载历史数据库
package components

import (
	"database/sql"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"

	fynexWidget "fyne.io/x/fyne/widget"
)

// 下载历史数据库 - 主 UI
func historyDatabase(_ fyne.Window) fyne.CanvasObject {
	historyDatabaseIcon, _ := fynexWidget.NewAnimatedGif(storage.NewFileURI("./img/historyDatabase.gif"))
	historyDatabaseIcon.SetMinSize(fyne.NewSize(290, 368))
	historyDatabaseIcon.Start()

	getHistoryButton := widget.NewButtonWithIcon("查看历史下载数据", theme.SearchIcon(),
		func() {
			errCode, _, areaList := queryArea()
			// 解析失败图片
			queryAreaFailedIcon := canvas.NewImageFromFile("./img/analyzeFailed.png")
			queryAreaFailedIcon.FillMode = canvas.ImageFillContain
			queryAreaFailedIcon.SetMinSize(fyne.NewSize(200, 200))

			var queryAreaResultWindow fyne.Window
			if errCode == 500 {
				tipFailed := widget.NewLabel("数据库查询失败，内部API错误")
				queryAreaResultWindow = fyne.CurrentApp().NewWindow("查询失败!")
				queryAreaResultWindow.SetContent(container.NewVBox(
					queryAreaFailedIcon,
					tipFailed,
					widget.NewButton("啊这!",
						func() { queryAreaResultWindow.Close() }),
				))
			} else if errCode == 404 {
				tipFailed := widget.NewLabel("无历史数据，请先下发下载任务")
				queryAreaResultWindow = fyne.CurrentApp().NewWindow("查询失败!")
				queryAreaResultWindow.SetContent(container.NewVBox(
					queryAreaFailedIcon,
					tipFailed,
					widget.NewButton("啊这!",
						func() { queryAreaResultWindow.Close() }),
				))
			} else if errCode == 200 {
				queryAreaResultWindow = fyne.CurrentApp().NewWindow("历史数据库")
				// fmt.Println(areaCount, areaList)
				queryAreaResultWindowTitle := widget.NewLabel("下面是历史数据库的查询结果，按照分区展示，可以点开查看游戏，点击游戏名称可以复制下载链接，在软件的“下发下载任务”处使用该链接下载")

				// 解析成功 - 列表组件
				// queryAccordionTabAreaData := make([]*widget.AccordionItem, areaCount)
				queryAccordionTab := widget.NewAccordion()

				// 页面可滚动
				// pageContainer := container.NewVBox(queryAreaResultWindowForm, queryAccordionTab)

				// 解析成功 - 表格展示组件
				queryAreaResultWindowForm := &widget.Form{
					Items: []*widget.FormItem{
						{Widget: queryAreaResultWindowTitle},
						{Text: "历史下载数据", Widget: queryAccordionTab},
					},
				}

				pageContainerScroll := container.NewHScroll(queryAreaResultWindowForm)

				for _, area := range areaList {
					currentItem := queryGameInfoByArea(area)
					queryAccordionTab.Append(currentItem)
				}

				queryAreaResultWindow.SetContent(pageContainerScroll)
				queryAreaResultWindow.Resize(fyne.NewSize(1280, 720))

			}
			queryAreaResultWindow.CenterOnScreen()
			queryAreaResultWindow.Show()
		},
	)

	searchHistoryButton := widget.NewButtonWithIcon("搜索历史下载数据", theme.SearchReplaceIcon(), func() {
		errCode, _, areaList := queryArea()
		// 解析失败图片
		queryAreaFailedIcon := canvas.NewImageFromFile("./img/analyzeFailed.png")
		queryAreaFailedIcon.FillMode = canvas.ImageFillContain
		queryAreaFailedIcon.SetMinSize(fyne.NewSize(200, 200))

		var queryAreaResultWindow fyne.Window

		queryResultContainer := container.NewVBox()

		if errCode == 500 {
			tipFailed := widget.NewLabel("数据库查询失败，内部API错误")
			queryAreaResultWindow = fyne.CurrentApp().NewWindow("查询失败!")
			queryAreaResultWindow.SetContent(container.NewVBox(
				queryAreaFailedIcon,
				tipFailed,
				widget.NewButton("啊这!",
					func() { queryAreaResultWindow.Close() }),
			))
		} else if errCode == 404 {
			tipFailed := widget.NewLabel("无历史数据，请先下发下载任务")
			queryAreaResultWindow = fyne.CurrentApp().NewWindow("查询失败!")
			queryAreaResultWindow.SetContent(container.NewVBox(
				queryAreaFailedIcon,
				tipFailed,
				widget.NewButton("啊这!",
					func() { queryAreaResultWindow.Close() }),
			))
		} else if errCode == 200 {
			queryAreaResultWindow = fyne.CurrentApp().NewWindow("搜索历史数据库")
			queryAreaResultWindowTitle := widget.NewLabel("选择分区，然后用 Galgame 的名称搜索，点击游戏名称可以复制下载链接，在软件的“下发下载任务”处使用该链接下载")
			resultBox := widget.NewLabelWithStyle("执行任务的结果显示在这里", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			var currentArea string

			areaSelector := widget.NewSelect(areaList, func(s string) {
				currentArea = s
			})

			searchNameEntry := widget.NewEntry()
			searchNameEntry.SetPlaceHolder("输入要搜索的 Gal 下载时保存的名称")

			searchByAreaButton := widget.NewButtonWithIcon("戳我按照选择的分区和名称(可选)搜索", theme.SearchIcon(), func() {
				var querySql string
				db, _ := sql.Open("sqlite3", "./yuukaDown.db")
				defer db.Close()
				if len(currentArea) < 1 {
					if len(searchNameEntry.Text) < 1 {
						resultBox.Text = "没有选择分区时必须输入 Gal 名称，什么都不输怎么搜索?!!你个阿露!!"
						resultBox.Refresh()
						return
					} else {
						resultBox.Text = "根据 Gal 名称搜索..."
						querySql = "SELECT galgameName, downloadBaseUrl, partNum, fileType FROM yuukaDownGalDB WHERE galgameName LIKE '%" + searchNameEntry.Text + "%';"
					}
				} else {
					if len(searchNameEntry.Text) < 1 {
						resultBox.Text = "根据分区搜索..."
						querySql = "SELECT galgameName, downloadBaseUrl, partNum, fileType FROM yuukaDownGalDB WHERE subArea='" + currentArea + "';"
					} else {
						resultBox.Text = "根据分区和 Gal 名称搜索..."
						querySql = "SELECT galgameName, downloadBaseUrl, partNum, fileType FROM yuukaDownGalDB WHERE galgameName LIKE '%" + searchNameEntry.Text + "%' AND subArea='" + currentArea + "';"
					}
				}

				// fmt.Println("querySql: ", querySql)
				queryGameInfoResult, err := db.Query(querySql)
				if err != nil {
					resultBox.Text += "搜索失败，错误原因: " + err.Error()
					resultBox.Refresh()
					return
				} else {
					queryResultContainer.RemoveAll()
					for queryGameInfoResult.Next() {
						var galgameName, downloadBaseUrl, partNum, fileType string
						queryGameInfoResult.Scan(&galgameName, &downloadBaseUrl, &partNum, &fileType)
						partNumInt, _ := strconv.ParseInt(partNum, 10, 64)
						var gameDownloadInfo string
						if partNumInt <= 1 {
							gameDownloadInfo = downloadBaseUrl + "." + fileType
						} else if partNumInt > 1 {
							gameDownloadInfo = downloadBaseUrl + ".part" + partNum + "." + fileType
						}
						gameInfoButton := widget.NewButton(galgameName, func() { clipboard.WriteAll(gameDownloadInfo) })
						queryResultContainer.Add(gameInfoButton)
					}
					queryResultContainer.Refresh()
					resultCount := len(queryResultContainer.Objects)
					resultBox.Text += "查询成功，数量:" + fmt.Sprint(resultCount)
					resultBox.Refresh()
				}
			})

			queryAreaResultWindowForm := &widget.Form{
				Items: []*widget.FormItem{
					{Widget: queryAreaResultWindowTitle},
					{Text: "选择分区(可选)", Widget: areaSelector},
					{Widget: searchNameEntry},
					{Widget: searchByAreaButton},
					{Text: "执行结果:", Widget: resultBox},
					{Text: "查询结果:", Widget: queryResultContainer},
				},
			}

			pageContainerScroll := container.NewHScroll(queryAreaResultWindowForm)

			queryAreaResultWindow.SetContent(pageContainerScroll)
			queryAreaResultWindow.Resize(fyne.NewSize(1280, 720))

		}
		queryAreaResultWindow.CenterOnScreen()
		queryAreaResultWindow.Show()

	})

	historyDatabaseFunction := widget.NewRadioGroup([]string{"开启"}, func(string) {})
	historyDatabaseFunction.SetSelected("开启")
	historyDatabaseFunction.Disable()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: historyDatabaseIcon},
			{Text: "历史数据库功能", Widget: historyDatabaseFunction},
			{Widget: getHistoryButton},
			{Widget: searchHistoryButton},
		},
	}

	return form

}

// 关于历史数据库的构思
// Step1 查询分类个数和内容->返回错误码，个数和内容
// 根据分类内容显示目录
// Step2 根据分类查询各分类下的游戏（分别按照不同类别查询）->返回最后一part下载链接和保存名saveName
// 根据返回生成按钮

// Step1 查询分类个数和内容->返回错误码，个数和内容
// errCode[200:"ok",404:"无任何值",500:"内部API错误"]
func queryArea() (errCode, areaCount int, areaList []string) {
	db, _ := sql.Open("sqlite3", "./yuukaDown.db")
	defer db.Close()
	// DISTINCT 去重，可选值理论上有 [game,krkr,rpg,3D] 几种
	queryAreaCountSql := "SELECT count(DISTINCT subArea) FROM yuukaDownGalDB;"
	queryAreaSql := "SELECT DISTINCT subArea FROM yuukaDownGalDB;"

	// 查询分类数量
	var areaCountNum int
	// areaCountNumQueryResult := db.QueryRow(queryAreaCountSql).Scan(&areaCountNum)
	db.QueryRow(queryAreaCountSql).Scan(&areaCountNum)
	if areaCountNum == 0 {
		errCode = 404
		return
	} else {
		areaCount = areaCountNum
	}

	// 查询分类条目
	queryAreaSqlResult, err := db.Query(queryAreaSql)
	if err != nil {
		errCode = 500
		return
	}
	for queryAreaSqlResult.Next() {
		var areaName string
		err = queryAreaSqlResult.Scan(&areaName)
		if err != nil {
			errCode = 500
			return
		}
		areaList = append(areaList, areaName)
	}
	errCode = 200
	return
}

// Step2 根据分类查询各分类下的游戏
func queryGameInfoByArea(areaName string) (accordionItem *widget.AccordionItem) {
	db, _ := sql.Open("sqlite3", "./yuukaDown.db")
	defer db.Close()

	querySql := "SELECT galgameName,downloadBaseUrl,partNum,fileType FROM yuukaDownGalDB WHERE subArea='" + areaName + "';"
	queryGameInfoByAreaResult, _ := db.Query(querySql)
	// var gameInfoContainer *fyne.Container
	gameInfoContainer := container.NewVBox()
	for queryGameInfoByAreaResult.Next() {
		var galgameName, downloadBaseUrl, partNum, fileType string
		queryGameInfoByAreaResult.Scan(&galgameName, &downloadBaseUrl, &partNum, &fileType)
		partNumInt, _ := strconv.ParseInt(partNum, 10, 64)
		var gameDownloadInfo string
		if partNumInt <= 1 {
			gameDownloadInfo = downloadBaseUrl + "." + fileType
		} else if partNumInt > 1 {
			gameDownloadInfo = downloadBaseUrl + ".part" + partNum + "." + fileType
		}
		gameInfoButton := widget.NewButton(galgameName, func() { clipboard.WriteAll(gameDownloadInfo) })
		gameInfoContainer.Add(gameInfoButton)
	}
	accordionItem = widget.NewAccordionItem(areaName, gameInfoContainer)
	return
}
