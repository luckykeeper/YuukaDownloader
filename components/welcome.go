package components

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func welcomeScreen(_ fyne.Window) fyne.CanvasObject {
	addImageIcon := canvas.NewImageFromFile("./img/yuuka.jpg")
	addImageIcon.FillMode = canvas.ImageFillContain
	addImageIcon.SetMinSize(fyne.NewSize(269.75, 480))

	return container.NewCenter(container.NewVBox(
		addImageIcon,
		widget.NewLabelWithStyle("↑优香酱可爱捏", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("yuukaDown, A software to Download Gal Through muti-Platform", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),

		container.NewHBox(
			widget.NewHyperlink("Powered By Luckykeeper", parseURL("https://luckykeeper.site/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("Github", parseURL("https://github.com/luckykeeper/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("Blog", parseURL("https://luckykeeper.site/")),
		),
		widget.NewLabel(""),
	))
}
