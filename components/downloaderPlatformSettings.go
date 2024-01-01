// 下载平台设定组件
package components

import (
	"database/sql"
	"fmt"
	"io"

	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/tidwall/gjson"
)

// 下载平台设定组件 - intro
func downloaderPlatformSettings(_ fyne.Window) fyne.CanvasObject {
	downloaderPlatformSettingsIcon := canvas.NewImageFromFile("./img/downloadPlatformConfig.jpg")
	downloaderPlatformSettingsIcon.FillMode = canvas.ImageFillContain
	downloaderPlatformSettingsIcon.SetMinSize(fyne.NewSize(580, 412))
	content := container.NewVBox(
		downloaderPlatformSettingsIcon,
		widget.NewLabel("在左侧的菜单中选择一个下载平台进行设置，记得保存和测试"))
	return container.NewCenter(content)
}

// 下载平台设定组件 - Aria2
func Aria2PlatformSetting(_ fyne.Window) fyne.CanvasObject {

	// 获取先前设定的参数
	var (
		// getConfigHaveResult                                             bool
		getConfigResult                               map[string]string
		aria2PlatformUrl, aria2jsonRPCVer, aria2Token string
	)

	log.Println("Aria2 Platform Selected")
	_, getConfigResult = getDownloadPlatformConfig("Aria2")
	// 不显密码
	if getConfigResult["ariaToken"] != "" {
		aria2Token = "**********"
	}
	aria2PlatformUrl = getConfigResult["platformUrl"]
	aria2jsonRPCVer = getConfigResult["jsonRPCVer"]

	Aria2PlatformSettingIcon := canvas.NewImageFromFile("./img/aria2Setting.jpg")
	Aria2PlatformSettingIcon.FillMode = canvas.ImageFillContain
	Aria2PlatformSettingIcon.SetMinSize(fyne.NewSize(580, 412))

	// Aria2 平台 UI

	inputPlatformUrl := widget.NewEntry()
	inputPlatformUrl.SetPlaceHolder("协议://地址(IP/域名):端口/jsonrpc")
	if aria2PlatformUrl != "" {
		inputPlatformUrl.SetText(aria2PlatformUrl)
	}

	inputAria2RPCVer := widget.NewEntry()
	inputAria2RPCVer.SetPlaceHolder("Aria2 一般填写 “2.0” ")
	if aria2jsonRPCVer != "" {
		inputAria2RPCVer.SetText(aria2jsonRPCVer)
	}

	inputAria2Token := widget.NewEntry()
	inputAria2Token.SetPlaceHolder("在这里填写 Aria2 的 RPC 密钥")
	if aria2Token != "" {
		inputAria2Token.SetPlaceHolder(aria2Token)
	}

	NoneInputTestErrorCode := widget.NewEntry()
	NoneInputTestErrorCode.SetPlaceHolder("点击测试后，这里将显示测试的结果")

	explainSaveErrorCode := widget.NewLabel("下面的文本框显示保存结果")
	explainTestErrorCode := widget.NewLabel("下面的文本框显示测试结果，200 为测试成功")

	savePlatformConfigBox := widget.NewEntry()
	savePlatformConfigBox.SetPlaceHolder("点击保存后，这里显示是否保存成功")

	saveButton := widget.NewButton("填好之后戳这里保存~（请勿重复点击）",
		func() {
			if inputPlatformUrl.Text != "" && inputAria2RPCVer.Text != "" && inputAria2Token.Text != "" {
				platformConfig := map[string]string{"PlatformUrl": inputPlatformUrl.Text, "Aria2RPCVer": inputAria2RPCVer.Text, "Aria2Token": inputAria2Token.Text}
				savePlatformSettings("Aria2", platformConfig)
				savePlatformConfigBox.SetPlaceHolder("保存成功!")
				// 填写成功后隐藏密码参数
				inputAria2Token.SetText("")
				inputAria2Token.SetPlaceHolder("**********")
				inputAria2Token.Refresh()
			} else {
				savePlatformConfigBox.SetPlaceHolder("保存失败，检查是否填写了所有参数")
			}
		},
	)

	testButton := widget.NewButton("然后点这里测试平台",
		func() {
			// 保存成功图片
			saveSucceedIcon := canvas.NewImageFromFile("./img/saveSuccess.jpg")
			saveSucceedIcon.FillMode = canvas.ImageFillContain
			saveSucceedIcon.SetMinSize(fyne.NewSize(162, 126))

			// 保存失败图片
			saveFailedIcon := canvas.NewImageFromFile("./img/saveFailed.jpg")
			saveFailedIcon.FillMode = canvas.ImageFillContain
			saveFailedIcon.SetMinSize(fyne.NewSize(200, 160))

			errorCode := testAria2()
			NoneInputTestErrorCode.SetPlaceHolder(errorCode)
			NoneInputTestErrorCode.Refresh()

			if errorCode == "200" {
				saveSuccessWindow := fyne.CurrentApp().NewWindow("测试成功!")
				saveSuccessWindow.SetContent(container.NewVBox(
					saveSucceedIcon,
					widget.NewLabel("你已经可以使用 Aria2 下载平台!"),
					widget.NewButton("好耶~",
						func() { saveSuccessWindow.Close() }),
				))

				saveSuccessWindow.CenterOnScreen()
				saveSuccessWindow.Show()
			} else {
				saveFailedWindow := fyne.CurrentApp().NewWindow("测试失败!")
				saveFailedWindow.SetContent(container.NewVBox(
					saveFailedIcon,
					widget.NewLabel("无法连接到 Aria2 下载平台，检查参数和网络!"),
					widget.NewButton("做不到就艾草!",
						func() { saveFailedWindow.Close() }),
				))

				saveFailedWindow.CenterOnScreen()
				saveFailedWindow.Show()
			}
		},
	)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: Aria2PlatformSettingIcon},
			{Text: "Aria2 服务端地址", Widget: inputPlatformUrl, HintText: "协议://地址(IP/域名):端口/jsonrpc"},
			{Text: "Aria2 RPC 版本", Widget: inputAria2RPCVer, HintText: "Aria2 一般填写 “2.0” "},
			{Text: "Aria2 服务端密钥", Widget: inputAria2Token, HintText: "在这里填写 Aria2 的 RPC 密钥"},

			{Widget: saveButton},
			{Widget: explainSaveErrorCode},
			{Text: "保存结果", Widget: savePlatformConfigBox, HintText: "点击保存后，这里显示是否保存成功"},

			{Widget: testButton},
			{Widget: explainTestErrorCode},
			{Text: "测试结果", Widget: NoneInputTestErrorCode, HintText: "测试平台连通性的结果"},
		},
	}

	return form

}

// 测试 Aria2 的连通性
func testAria2() (errorCode string) {
	_, getConfigResult := getDownloadPlatformConfig("Aria2")
	aria2PlatformUrl := getConfigResult["platformUrl"]
	aria2jsonRPCVer := getConfigResult["jsonRPCVer"]
	aria2Token := getConfigResult["ariaToken"]

	requestID := generateRandomID()

	method := "POST"

	payload := strings.NewReader("{" + `"jsonrpc":"` + aria2jsonRPCVer + `",` + `"id":"` + requestID + `","method":"aria2.getGlobalStat",` + `"params": ["token:` + aria2Token + `"]}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, aria2PlatformUrl, payload)

	if err != nil {
		errorCode = err.Error()
		return
	}

	res, err := client.Do(req)
	if err != nil {
		errorCode = err.Error()
		return
	}
	defer res.Body.Close()
	errorCode = fmt.Sprint(res.StatusCode)
	return
}

// 获取 Aria2 状态
func testAria2Status() (errorCode, downloadSpeed, uploadSpeed, numActive, numWaiting string) {
	_, getConfigResult := getDownloadPlatformConfig("Aria2")
	aria2PlatformUrl := getConfigResult["platformUrl"]
	aria2jsonRPCVer := getConfigResult["jsonRPCVer"]
	aria2Token := getConfigResult["ariaToken"]

	requestID := generateRandomID()

	method := "POST"

	payload := strings.NewReader("{" + `"jsonrpc":"` + aria2jsonRPCVer + `",` + `"id":"` + requestID + `","method":"aria2.getGlobalStat",` + `"params": ["token:` + aria2Token + `"]}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, aria2PlatformUrl, payload)

	if err != nil {
		errorCode = err.Error()
		return
	}

	res, err := client.Do(req)
	if err != nil {
		errorCode = err.Error()
		return
	}
	defer res.Body.Close()
	aria2StatusResponse, _ := io.ReadAll(res.Body)
	log.Println(string(aria2StatusResponse))
	errorCode = fmt.Sprint(res.StatusCode)
	downloadSpeed = fmt.Sprint(gjson.Get(string(aria2StatusResponse), "result.downloadSpeed"))
	uploadSpeed = fmt.Sprint(gjson.Get(string(aria2StatusResponse), "result.uploadSpeed"))
	numActive = fmt.Sprint(gjson.Get(string(aria2StatusResponse), "result.numActive"))
	numWaiting = fmt.Sprint(gjson.Get(string(aria2StatusResponse), "result.numWaiting"))
	// 单位是 Byte/s ，除以两个1024的乘积
	downloadSpeedInt64, _ := strconv.ParseInt(downloadSpeed, 10, 64)
	uploadSpeedInt64, _ := strconv.ParseInt(uploadSpeed, 10, 64)
	downloadSpeed = fmt.Sprint(downloadSpeedInt64 / 1048576)
	uploadSpeed = fmt.Sprint(uploadSpeedInt64 / 1048576)
	return
}

// 生成随机 ID
func generateRandomID() (id string) {
	// 当前时间，精确到纳秒
	getNowTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(),
		time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(), time.Local).Format("2006-01-02 15:04:05")
	getNowTime = strings.Replace(getNowTime, " ", "", -1)
	getNowTime = strings.Replace(getNowTime, "-", "", -1)
	getNowTime = strings.Replace(getNowTime, ":", "", -1)

	// id 由当前时间+ 32 位随机字符串构成
	id = getNowTime + randomStr(32)
	return
}

// 生成随机字符串，降低撞车概率
func randomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
	for i := 0; i < length; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

// 判断下载平台参数是否存在
func getDownloadPlatformConfig(platform string) (haveResult bool, result map[string]string) {
	db, _ := sql.Open("sqlite3", "./yuukaDown.db")
	defer db.Close()
	if platform == "Aria2" {
		log.Printf("Getting Aria2 Config...")
		querySql := "select platformUrl,jsonRPCVer,ariaToken from yuukaDownPlatform where downloaderPlatform='" + platform + "';"
		var data, data1, data2 string
		queryResult := db.QueryRow(querySql).Scan(&data, &data1, &data2)
		if queryResult == sql.ErrNoRows {
			haveResult = false
			log.Println("done!")
			return
		} else { // 之前设置过相关参数
			haveResult = true
			result = make(map[string]string)
			result["platformUrl"] = data
			result["jsonRPCVer"] = data1
			result["ariaToken"] = data2
			log.Println("done!")
			return
		}
	} else {
		haveResult = false
		log.Println("done!")
		return
	}

}

// 保存下载平台参数到数据库
func savePlatformSettings(platform string, platformConfig map[string]string) {
	db, _ := sql.Open("sqlite3", "./yuukaDown.db")
	defer db.Close()
	if platform == "Aria2" {
		platformUrl := platformConfig["PlatformUrl"]
		jsonRPCVer := platformConfig["Aria2RPCVer"]
		ariaToken := platformConfig["Aria2Token"]
		db.Exec("UPDATE yuukaDownPlatform SET platformUrl='" + platformUrl + "',jsonRPCVer='" + jsonRPCVer + "',ariaToken='" + ariaToken + "' WHERE downloaderPlatform='" + platform + "';")
	}
}
