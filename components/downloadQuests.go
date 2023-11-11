// 下载平台任务下发组件
package components

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/atotto/clipboard"
)

// 下载平台任务下发组件 - 主 UI
func downloadQuests(_ fyne.Window) fyne.CanvasObject {
	downloadLinkBox := widget.NewEntry()
	downloadLinkBox.SetPlaceHolder("复制最后part链接至此")

	manualPrefixBox := widget.NewEntry()
	manualPrefixBox.SetPlaceHolder("其它前缀在此输入")

	fufugalPrefixBox := container.NewHBox(
		widget.NewButtonWithIcon("", theme.ContentPasteIcon(),
			func() {
				downloadLinkBox.SetText(readClipboard())
				downloadLinkBox.Refresh()
			},
		),
		widget.NewButton("解析",
			func() { analyzeButton("", downloadLinkBox) },
		),
		widget.NewLabel("切换下载站"),
		widget.NewButton("@",
			func() { analyzeButton("@", downloadLinkBox) },
		),
		widget.NewButton("zz",
			func() { analyzeButton("zz", downloadLinkBox) },
		),
		widget.NewButton("qq",
			func() { analyzeButton("qq", downloadLinkBox) },
		),
		widget.NewButton("gs",
			func() { analyzeButton("gs", downloadLinkBox) },
		),
		widget.NewButton("手动解析",
			func() { analyzeButton(manualPrefixBox.Text, downloadLinkBox) },
		),
		widget.NewLabel("若前缀不为前面选项时，可以在下面的输入框手动输入前缀并手动解析（举例：https://qq.llgal.xyz 的前缀为“qq”）"),
	)

	aria2SettingStatus := widget.NewLabel("testing...")
	aria2ConnStatus := widget.NewLabel("testing...")
	aria2DownStatus := widget.NewLabel("testing...")
	aria2UploadStatus := widget.NewLabel("testing...")
	aria2NumActive := widget.NewLabel("testing...")
	aria2NumWaiting := widget.NewLabel("testing...")

	// 修改：展示 Aria2 连通性和参数
	aria2StatusBox := container.NewHBox(
		widget.NewLabel("Aria2——设置状态："),
		aria2SettingStatus,
		widget.NewLabel("，连通："),
		aria2ConnStatus,
		widget.NewLabel("，下载速度："),
		aria2DownStatus,
		widget.NewLabel("M/s，上传速度"),
		aria2UploadStatus,
		widget.NewLabel("M/s，正在下载任务数："),
		aria2NumActive,
		widget.NewLabel("，等待下载任务数："),
		aria2NumWaiting,
	)

	// 刷新显示 Aria2 连通性和参数的函数
	go func() {
		for range time.Tick(time.Second * 5) {
			haveResult, _ := getDownloadPlatformConfig("Aria2")
			if haveResult {
				aria2SettingStatus.SetText("True")
				aria2ConnStatus.Refresh()
				if errorCode, downloadSpeed, uploadSpeed, numActive, numWaiting := testAria2Status(); errorCode == "200" {
					aria2ConnStatus.SetText("True")
					aria2ConnStatus.Refresh()

					aria2DownStatus.SetText(downloadSpeed)
					aria2DownStatus.Refresh()

					aria2UploadStatus.SetText(uploadSpeed)
					aria2UploadStatus.Refresh()

					aria2NumActive.SetText(numActive)
					aria2NumActive.Refresh()

					aria2NumWaiting.SetText(numWaiting)
					aria2NumWaiting.Refresh()
				} else {
					aria2ConnStatus.SetText("False")
					aria2ConnStatus.Refresh()

					aria2DownStatus.SetText("testing...")
					aria2DownStatus.Refresh()

					aria2UploadStatus.SetText("testing...")
					aria2UploadStatus.Refresh()

					aria2NumActive.SetText("testing...")
					aria2NumActive.Refresh()

					aria2NumWaiting.SetText("testing...")
					aria2NumWaiting.Refresh()
				}
			} else {
				aria2SettingStatus.SetText("False")
				aria2ConnStatus.Refresh()

				aria2ConnStatus.SetText("False")
				aria2ConnStatus.Refresh()

				aria2DownStatus.SetText("testing...")
				aria2DownStatus.Refresh()

				aria2UploadStatus.SetText("testing...")
				aria2UploadStatus.Refresh()

				aria2NumActive.SetText("testing...")
				aria2NumActive.Refresh()

				aria2NumWaiting.SetText("testing...")
				aria2NumWaiting.Refresh()
			}
		}
	}()

	// ContentTitle = widget.NewLabel("下发下载任务到平台~")
	// ContentTitleBox = container.New(layout.NewVBoxLayout(), ContentTitle, widget.NewSeparator())

	downloadQuestsIcon := canvas.NewImageFromFile("./img/downloadQuests.jpg")
	downloadQuestsIcon.FillMode = canvas.ImageFillContain
	downloadQuestsIcon.SetMinSize(fyne.NewSize(240, 339))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: downloadQuestsIcon},
			{Text: "下载任务链接", Widget: downloadLinkBox, HintText: "复制最后一Part（无part的直接复制），程序会自动解析其它part | 可以使用下面的粘贴按钮快速完成复制"},
			{Widget: fufugalPrefixBox},
			{Text: "手动解析前缀", Widget: manualPrefixBox, HintText: "输入完成后点击“手动解析”"},
			// {Widget: downloadQuestsBox},
			{Widget: aria2StatusBox},
		},
	}

	return form

}

// 下发下载任务到 Aria2 平台
func aria2DownloadQuest(downloadBaseUrl, partNum, fileType string) (success, fault int) {
	success = 0
	fault = 0

	_, getConfigResult := getDownloadPlatformConfig("Aria2")
	aria2PlatformUrl := getConfigResult["platformUrl"]
	aria2jsonRPCVer := getConfigResult["jsonRPCVer"]
	aria2Token := getConfigResult["ariaToken"]

	requestID := generateRandomID()

	method := "POST"
	partNumInt64, _ := strconv.ParseInt(partNum, 10, 64)
	for i := 0; i < int(partNumInt64); i++ {
		var payload *strings.Reader
		if int(partNumInt64) > 1 {
			payload = strings.NewReader("{" + `"jsonrpc":"` + aria2jsonRPCVer + `",` + `"id":"` + requestID + `","method":"aria2.addUri",` + `"params": ["token:` + aria2Token + `",` +
				`["` + downloadBaseUrl + ".part" + strconv.Itoa(i+1) + "." + fileType + `"]` + `]}`)
		} else {
			payload = strings.NewReader("{" + `"jsonrpc":"` + aria2jsonRPCVer + `",` + `"id":"` + requestID + `","method":"aria2.addUri",` + `"params": ["token:` + aria2Token + `",` +
				`["` + downloadBaseUrl + "." + fileType + `"]` + `]}`)
		}
		client := &http.Client{}
		req, err := http.NewRequest(method, aria2PlatformUrl, payload)

		if err != nil {
			fault = fault + 1
		}

		res, err := client.Do(req)
		if err != nil {
			fault = fault + 1
			return
		}
		defer res.Body.Close()
		errorCode := fmt.Sprint(res.StatusCode)
		if errorCode != "200" {
			fault = fault + 1
		} else {
			success = success + 1
		}
	}
	return
}

// 解析按钮
func analyzeButton(usePrefix string, downloadLinkBox *widget.Entry) {
	// 解析失败图片
	analyzeFailedIcon := canvas.NewImageFromFile("./img/analyzeFailed.png")
	analyzeFailedIcon.FillMode = canvas.ImageFillContain
	analyzeFailedIcon.SetMinSize(fyne.NewSize(200, 200))

	// Aria2 下发下载任务结果页面图片
	aria2DownloadQuestResultIcon := canvas.NewImageFromFile("./img/aria2DownloadQuestResult.jpg")
	aria2DownloadQuestResultIcon.FillMode = canvas.ImageFillContain
	aria2DownloadQuestResultIcon.SetMinSize(fyne.NewSize(216, 305.8))

	analyzedResult := make(map[string]string)
	var errCode int
	analyzedResult, errCode = analyzeUrl(usePrefix, downloadLinkBox.Text)
	var analyzeResultWindow fyne.Window
	if errCode != 200 {
		var tipFailed *widget.Label
		if errCode == 400 {
			tipFailed = widget.NewLabel("解析失败，疑似非Url链接？")
		} else if errCode == 406 {
			tipFailed = widget.NewLabel("解析失败，疑似非初音站链接？")
		} else {
			tipFailed = widget.NewLabel("解析失败，疑似是程序问题？找作者修Bug去！")
		}
		analyzeResultWindow = fyne.CurrentApp().NewWindow("解析失败!")
		analyzeResultWindow.SetContent(container.NewVBox(
			analyzeFailedIcon,
			tipFailed,
			widget.NewButton("啊这!",
				func() { analyzeResultWindow.Close() }),
		))
	} else {
		// 解析成功 - 基础组件
		tipSuccess := widget.NewLabel("解析成功！请选择下载信息并发送平台，或点击右上角“X”放弃下载")

		fileNameUnescape, _ := url.QueryUnescape(analyzedResult["fileName"])
		folderNameUnescape, _ := url.QueryUnescape(analyzedResult["folderName"])

		var saveName string
		chooseSaveName := widget.NewRadioGroup([]string{fileNameUnescape, folderNameUnescape}, func(userChoice string) {
			saveName = userChoice
			// 对 SQL 不能插入的字符进行替换
			saveName = strings.Replace(saveName, "\"", "”", -1)
			saveName = strings.Replace(saveName, "'", "’", -1)
			log.Println("User choose savename:", saveName)
		})
		chooseSaveName.Horizontal = false
		chooseSaveName.SetSelected(fileNameUnescape)

		fileTotalTips0 := widget.NewLabel("文件相关信息：")
		fileTotalTips1 := widget.NewLabel("文件名称：" + fileNameUnescape + "，文件夹名称：" + folderNameUnescape)
		fileTotalTips2 := widget.NewLabel("分区：" + analyzedResult["subUrl"] + "，part数量：" + analyzedResult["partNum"])
		fileTotalTips3 := widget.NewLabel("下载线路：" + analyzedResult["prefix"] + "，文件后缀：" + analyzedResult["fileType"])
		fileTotalTips4 := widget.NewLabel("下载任务预览：")

		// 提交到 Aria2 下载平台
		downloadQuestThroughAria2Button := widget.NewButton("下发任务到 Aria2 平台",
			func() {
				success, fault := aria2DownloadQuest(analyzedResult["downloadBaseUrl"], analyzedResult["partNum"], analyzedResult["fileType"])
				saveToHistoryDatabe(saveName, analyzedResult["downloadBaseUrl"], analyzedResult["partNum"], analyzedResult["fileType"], analyzedResult["subUrl"])

				aria2DownloadQuestResultWindow := fyne.CurrentApp().NewWindow("Aria2 - 下发任务结果")
				aria2DownloadQuestResultWindow.SetContent(container.NewVBox(
					aria2DownloadQuestResultIcon,
					widget.NewLabel("下发了"+analyzedResult["partNum"]+"个下载任务。成功"+fmt.Sprint(success)+"个，失败"+fmt.Sprint(fault)+"个~"),
					widget.NewButton("太好玩啦!",
						func() {
							aria2DownloadQuestResultWindow.Close()
							analyzeResultWindow.Close()
						}),
				))
				aria2DownloadQuestResultWindow.CenterOnScreen()
				aria2DownloadQuestResultWindow.Show()
			})

		if errCode := testAria2(); errCode != "200" {
			downloadQuestThroughAria2Button.Disable()
		} else if errCode == "200" && downloadQuestThroughAria2Button.Disabled() {
			downloadQuestThroughAria2Button.Enable()
		}

		chooseDownloadPlatform := widget.NewRadioGroup([]string{"Aria2"}, func(string) {})
		chooseDownloadPlatform.SetSelected("Aria2")
		chooseDownloadPlatform.Disable()

		// 解析成功 - 列表组件
		analyzeSuccessListDataInt64, _ := strconv.ParseInt(analyzedResult["partNum"], 10, 64)
		analyzeSuccessListData := make([]string, analyzeSuccessListDataInt64)
		for i := range analyzeSuccessListData {
			if analyzeSuccessListDataInt64 > 1 {
				analyzeSuccessListData[i] = analyzedResult["downloadBaseUrl"] + ".part" + strconv.Itoa(i+1) + "." + analyzedResult["fileType"]
			} else {
				analyzeSuccessListData[i] = analyzedResult["downloadBaseUrl"] + "." + analyzedResult["fileType"]
			}
		}

		analyzeSuccessList := widget.NewList(

			func() int {
				return len(analyzeSuccessListData)
			},
			func() fyne.CanvasObject {
				return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template Object"))
			},
			func(id widget.ListItemID, item fyne.CanvasObject) {
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(analyzeSuccessListData[id])
			},
		)
		analyzeSuccessList.SetItemHeight(5, 50)

		// 解析成功 - 表格展示组件
		analyzeSuccessForm := &widget.Form{
			Items: []*widget.FormItem{
				{Widget: tipSuccess},
			},
		}
		analyzeSuccessForm.Append("选择保存到软件本地数据库的名称", chooseSaveName)
		analyzeSuccessForm.Append("选择下载平台", chooseDownloadPlatform)
		analyzeSuccessForm.Append("", fileTotalTips0)
		analyzeSuccessForm.Append("", fileTotalTips1)
		analyzeSuccessForm.Append("", fileTotalTips2)
		analyzeSuccessForm.Append("", fileTotalTips3)
		analyzeSuccessForm.Append("", downloadQuestThroughAria2Button)
		analyzeSuccessForm.Append("", fileTotalTips4)
		// analyzeSuccessForm.Append("", analyzeSuccessList)

		analyzeResultWindow = fyne.CurrentApp().NewWindow("解析成功!")
		// analyzeResultWindow.SetContent(analyzeSuccessForm)
		analyzeResultWindow.SetContent(container.NewVSplit(analyzeSuccessForm, analyzeSuccessList))
		analyzeResultWindow.Resize(fyne.NewSize(1024, 768))
	}
	analyzeResultWindow.CenterOnScreen()
	analyzeResultWindow.Show()
}

// 保存到历史数据库
func saveToHistoryDatabe(saveName, downloadBaseUrl, partNum, fileType, subUrl string) {
	db, _ := sql.Open("sqlite3", "./yuukaDown.db")
	log.Println("saveName:", saveName, ",downloadBaseUrl:", downloadBaseUrl, ",partNum:", partNum, ",fileType:", fileType, ",subUrl:", subUrl)
	defer db.Close()

	log.Printf("Saving To History Database...")
	querySql := "select galgameName from yuukaDownGalDB where galgameName='" + saveName + "';"
	var data string
	queryResult := db.QueryRow(querySql).Scan(&data)
	if queryResult == sql.ErrNoRows {
		insertSql := "insert into yuukaDownGalDB (galgameName,downloadBaseUrl,partNum,fileType,subArea) VALUES ('" + saveName + "'," + "'" + downloadBaseUrl + "','" + partNum + "'," + "'" + fileType + "'," + "'" + subUrl + "');"
		log.Println("将执行SQL：", insertSql)
		db.Exec(insertSql)
		log.Println("inserting...done!")
	} else { // 之前设置过相关参数
		log.Printf("Date Exists, Updating Database...")
		updateSql := "update yuukaDownGalDB SET galgameName='" + saveName + "',downloadBaseUrl='" + downloadBaseUrl + "',partNum='" + partNum + "',fileType='" + fileType + "',subArea='" + subUrl + "' WHERE galgameName='" + saveName + "';"
		log.Println("将执行SQL：", updateSql)
		db.Exec(updateSql)
		log.Println("done!")
	}
}

// 读取剪贴板
func readClipboard() (clipboardText string) {
	clipboardText, err := clipboard.ReadAll()
	if err != nil {
		return "剪贴板内不是文字!"
	} else {
		return clipboardText
	}
}

// 解析下载链接
func analyzeUrl(usePrefix, urlToBeAnalyze string) (analyzedResult map[string]string, errCode int) {
	log.Println("送解析Url:", urlToBeAnalyze, "!")
	// Url 示例（多p）：https://zz.llgal.xyz/game/%E9%AD%94%E5%A5%B3%E7%9A%84%E8%8A%B1%E5%9B%AD/Witch%27s%20Garden.part4.rar
	// Url 示例（单文件）：https://llgal.xyz/rpg/%E5%AE%B3%E7%BE%9E%E7%9A%84%E6%A4%8E%E5%90%8D%E9%85%B1/HazukaShina-chan%20v1.0.0%20ZH.rar

	// To Be Fix：多层级文件夹的名称解析问题，示例如下
	// 思路：从文件夹按照倒序取，先取后缀，再取part，再取名称，再取上一级的文件夹
	// https://qq.llgal.xyz/krkr/%E6%97%A7/krkr/%E7%BC%A4%E7%BA%B7%E5%B0%91%E5%A5%B3%20colorful%20cure/%E7%BC%A4%E7%BA%B7%E5%B0%91%E5%A5%B3%20colorful%20cure.rar

	// 下面的返回数据要加一个完整的Url（但是不带part和后缀，由partNum判断），供下载使用【判断是初音站链接（errCode=200）后给出】
	// 返回数据示例：["protocol":"https","prefix":"zz","subUrl":"game","partNum":"4","folderName":"4","fileName":"害羞的椎名酱","fileType":"rar",
	// "downloadBaseUrl":"https://llgal.xyz/rpg/%E5%AE%B3%E7%BE%9E%E7%9A%84%E6%A4%8E%E5%90%8D%E9%85%B1/HazukaShina-chan%20v1.0.0%20ZH"],200
	// errCode:[200:"ok",400:"非URL链接",406:"非初音站链接"]
	analyzedResult = make(map[string]string)

	// Step1 提取链接头
	// e.g. https
	protocol := strings.Split(urlToBeAnalyze, "://")[0]
	if protocol != "https" && protocol != "http" {
		log.Println("疑似非Url链接")
		return nil, 400
	}
	if len(strings.Split(urlToBeAnalyze, "llgal.xyz/")) < 2 {
		return nil, 406
	}
	log.Println("协议：", protocol)
	urlToBeAnalyze = strings.Split(urlToBeAnalyze, "://")[1]
	log.Println("余下部分：", urlToBeAnalyze)

	analyzedResult["protocol"] = protocol

	// Step2 提取前缀
	// e.g. zz
	prefix := strings.Split(urlToBeAnalyze, "llgal.xyz/")[0]
	// 给定前缀时使用指定前缀，否则进行分析
	if usePrefix != "" {
		if usePrefix == "@" {
			analyzedResult["prefix"] = ""
		} else {
			analyzedResult["prefix"] = usePrefix
		}
	} else {
		if prefix != "" && prefix != "zz." && prefix != "qq." && prefix != "gs." {
			log.Println("疑似非初音站链接")
			return nil, 406
		}
		// 有一级域名时去掉“.”，“@”域名直接赋值
		if len(prefix) > 0 {
			analyzedResult["prefix"] = prefix[:len(prefix)-1]
			log.Println("前缀：", prefix[:len(prefix)-1])
		} else {
			analyzedResult["prefix"] = prefix
			log.Println("前缀：", prefix)
		}
	}
	urlToBeAnalyze = strings.Split(urlToBeAnalyze, "llgal.xyz/")[1]
	log.Println("余下部分：", urlToBeAnalyze)

	// Step3 提取子链接
	// e.g. game/
	subUrl := strings.Split(urlToBeAnalyze, "/")[0]
	log.Println("子链接：", subUrl)
	analyzedResult["subUrl"] = subUrl

	// Step4 提取格式
	toBeCheck := strings.Split(urlToBeAnalyze, ".")
	urlToBeAnalyze = ""
	log.Println("文件类型：", toBeCheck[len(toBeCheck)-1])
	analyzedResult["fileType"] = toBeCheck[len(toBeCheck)-1]
	for i := 0; i < len(toBeCheck)-1; i++ {
		if len(urlToBeAnalyze) > 0 {
			urlToBeAnalyze = urlToBeAnalyze + "." + toBeCheck[i]
		} else {
			urlToBeAnalyze = urlToBeAnalyze + toBeCheck[i]
		}
	}
	log.Println("余下部分：", urlToBeAnalyze)

	// Step5 提取 part 数量
	toBeCheck = strings.Split(urlToBeAnalyze, ".part")
	if len(toBeCheck) > 1 {
		log.Println("检测到多 part!")
		analyzedResult["partNum"] = toBeCheck[len(toBeCheck)-1]
		log.Println("part数：", toBeCheck[len(toBeCheck)-1])

		// 对字符串进行了分割，更新剩下部分
		urlToBeAnalyze = ""
		for i := 0; i < len(toBeCheck)-1; i++ {
			if len(urlToBeAnalyze) > 0 {
				urlToBeAnalyze = urlToBeAnalyze + "." + toBeCheck[i]
			} else {
				urlToBeAnalyze = urlToBeAnalyze + toBeCheck[i]
			}
		}
	} else {
		// 未对字符串进行分割，无需更新剩下部分
		log.Println("未检测到多 part!")
		analyzedResult["partNum"] = "1"
		log.Println("part数：1")
	}
	log.Println("余下部分：", urlToBeAnalyze)

	// Step6 返回基本 Url
	if len(analyzedResult["prefix"]) > 0 {
		analyzedResult["downloadBaseUrl"] = analyzedResult["protocol"] + "://" + analyzedResult["prefix"] + ".llgal.xyz/" + urlToBeAnalyze
	} else {
		analyzedResult["downloadBaseUrl"] = analyzedResult["protocol"] + "://llgal.xyz/" + urlToBeAnalyze
	}

	// Step7 提取文件名称
	toBeCheck = strings.Split(urlToBeAnalyze, "/")
	analyzedResult["fileName"] = toBeCheck[len(toBeCheck)-1]
	log.Println("文件名（编码）：", analyzedResult["fileName"])
	if fileNameUnescape, err := url.QueryUnescape(analyzedResult["fileName"]); err != nil {
		log.Println("疑似非初音站链接")
		return nil, 406
	} else {
		log.Println("文件名（解码）：", fileNameUnescape)
	}

	urlToBeAnalyze = ""
	for i := 0; i < len(toBeCheck)-1; i++ {
		if len(urlToBeAnalyze) > 0 {
			urlToBeAnalyze = urlToBeAnalyze + "/" + toBeCheck[i]
		} else {
			urlToBeAnalyze = urlToBeAnalyze + toBeCheck[i]
		}
	}
	// Step8 提取文件夹名称（只提取上一级的文件夹）
	toBeCheck = strings.Split(urlToBeAnalyze, "/")
	analyzedResult["folderName"] = toBeCheck[len(toBeCheck)-1]
	log.Println("文件夹名（编码）：", analyzedResult["folderName"])
	if folderNameUnescape, err := url.QueryUnescape(analyzedResult["folderName"]); err != nil {
		log.Println("疑似非初音站链接")
		return nil, 406
	} else {
		log.Println("文件夹名（解码）：", folderNameUnescape)
	}

	log.Println("analyzedResult:", analyzedResult)
	errCode = 200
	return

	// 20230420 修改为上面的倒序获取
	// // Step4 提取文件夹名称
	// folderName := strings.Split(urlToBeAnalyze, "/")[1]
	// log.Println("文件夹名称（编码）：", folderName)
	// analyzedResult["folderName"] = folderName

	// if folderNameUnescape, err := url.QueryUnescape(folderName); err != nil {
	// 	log.Println("疑似非初音站链接")
	// 	return nil, 406
	// } else {
	// 	log.Println("文件夹名称（解码）：", folderNameUnescape)
	// 	urlToBeAnalyze = strings.Split(urlToBeAnalyze, "/")[2]
	// 	log.Println("余下部分：", urlToBeAnalyze)

	// 	// Step5 提取文件名
	// 	// e.g. HazukaShina-chan v1.0.0 ZH（话说这玩意儿是真的难打，通个关太难了...）
	// 	fileNameArr := strings.Split(urlToBeAnalyze, ".part")
	// 	var fileName string
	// 	if len(fileNameArr) > 1 {
	// 		log.Println("检测到多 part!")
	// 		fileName = fileNameArr[0]
	// 		analyzedResult["fileName"] = fileName
	// 		fileNameLast := fileNameArr[1]
	// 		log.Println("文件名称（编码）：", fileName)
	// 		if fileNameUnescape, err := url.QueryUnescape(fileName); err != nil {
	// 			log.Println("疑似非初音站链接")
	// 			return nil, 406
	// 		} else {
	// 			log.Println("文件名称（解码）：", fileNameUnescape)
	// 		}
	// 		log.Println("part:", strings.Split(fileNameLast, ".")[0])
	// 		log.Println("fileType:", strings.Split(fileNameLast, ".")[1])
	// 		analyzedResult["part"] = strings.Split(fileNameLast, ".")[0]
	// 		analyzedResult["fileType"] = strings.Split(fileNameLast, ".")[1]

	// 	} else {
	// 		fileName = strings.Split(urlToBeAnalyze, ".rar")[0]
	// 		analyzedResult["fileName"] = fileName
	// 		if fileNameUnescape, err := url.QueryUnescape(fileName); err != nil {
	// 			log.Println("疑似非初音站链接")
	// 			return nil, 406
	// 		} else {
	// 			log.Println("文件名称（解码）：", fileNameUnescape)
	// 		}
	// 		// 暂且推定均为 rar 格式
	// 		log.Println("fileType:", "rar")
	// 		analyzedResult["part"] = "1"
	// 		analyzedResult["fileType"] = "rar"
	// 	}
	// 	errCode = 200
	// 	log.Println("func - analyzeUrl:")
	// 	log.Println("analyzedResult:", analyzedResult)
	// 	log.Println("errCode:", errCode)
	// 	log.Println("_____________")
	// 	return
	// }
}
