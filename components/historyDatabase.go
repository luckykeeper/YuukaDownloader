// 下载历史数据库
package components

import (
	"database/sql"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"

	fynexWidget "fyne.io/x/fyne/widget"
)

// 下载历史数据库 - 主 UI
func historyDatabase(_ fyne.Window) fyne.CanvasObject {
	historyDatabaseIcon, _ := fynexWidget.NewAnimatedGif(storage.NewFileURI("./img/historyDatabase.gif"))
	historyDatabaseIcon.SetMinSize(fyne.NewSize(290, 368))
	historyDatabaseIcon.Start()

	getHistoryButton := widget.NewButton("查看历史下载数据",
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

				// 解析成功 - 表格展示组件
				queryAreaResultWindowForm := &widget.Form{
					Items: []*widget.FormItem{
						{Widget: queryAreaResultWindowTitle},
					},
				}

				// 解析成功 - 列表组件
				// queryAccordionTabAreaData := make([]*widget.AccordionItem, areaCount)
				queryAccordionTab := widget.NewAccordion()

				// 页面可滚动
				pageContainer := container.NewVBox(queryAreaResultWindowForm, queryAccordionTab)
				pageContainerScroll := container.NewHScroll(pageContainer)

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

	historyDatabaseFunction := widget.NewRadioGroup([]string{"开启"}, func(string) {})
	historyDatabaseFunction.SetSelected("开启")
	historyDatabaseFunction.Disable()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: historyDatabaseIcon},
			{Text: "历史数据库功能", Widget: historyDatabaseFunction},
			{Widget: getHistoryButton},
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
