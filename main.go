package main

import (
	"image/color"
	"log"
	"os"

	"eliasnaur.com/font/roboto/robotobold"
	"eliasnaur.com/font/roboto/robotoregular"
	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	expl "gioui.org/x/explorer"
	"github.com/g45t345rt/g45w/app_instance"
	"github.com/g45t345rt/g45w/containers/bottom_bar"
	"github.com/g45t345rt/g45w/containers/node_status_bar"
	"github.com/g45t345rt/g45w/containers/notification_modals"
	"github.com/g45t345rt/g45w/integrated_node"
	"github.com/g45t345rt/g45w/node_manager"
	page_node "github.com/g45t345rt/g45w/pages/node"
	page_settings "github.com/g45t345rt/g45w/pages/settings"
	page_wallet "github.com/g45t345rt/g45w/pages/wallet"
	page_wallet_select "github.com/g45t345rt/g45w/pages/wallet_select"
	"github.com/g45t345rt/g45w/router"
	"github.com/g45t345rt/g45w/settings"
	"github.com/g45t345rt/g45w/utils"
	"github.com/g45t345rt/g45w/wallet_manager"
)

func loadFontCollection() ([]font.FontFace, error) {
	robotoRegular, err := opentype.Parse(robotoregular.TTF)
	if err != nil {
		return nil, err
	}

	robotoBold, err := opentype.Parse(robotobold.TTF)
	if err != nil {
		return nil, err
	}

	fontCollection := []font.FontFace{}
	//gofont.Collection()
	fontCollection = append(fontCollection, font.FontFace{Font: font.Font{}, Face: robotoRegular})
	fontCollection = append(fontCollection, font.FontFace{Font: font.Font{Weight: font.Bold}, Face: robotoBold})
	return fontCollection, nil
}

func runApp() error {
	var ops op.Ops

	window := app_instance.Window
	th := app_instance.Theme
	router := app_instance.Router
	explorer := app_instance.Explorer

	for {
		e := <-window.Events()
		explorer.ListenEvents(e)
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			router.Layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

func main() {
	err := settings.Instantiate().Load()
	if err != nil {
		log.Fatal(err)
	}

	err = wallet_manager.Instantiate().Load()
	if err != nil {
		log.Fatal(err)
	}

	integrated_node.Instantiate()

	err = node_manager.Instantiate().Load()
	if err != nil {
		log.Fatal(err)
	}

	// window
	minSizeX := unit.Dp(375)
	minSizeY := unit.Dp(600)
	maxSizeX := unit.Dp(500)
	maxSizeY := unit.Dp(1000)

	window := app.NewWindow(
		app.Title("G45W"),
		app.MinSize(minSizeX, minSizeY),
		app.Size(minSizeX, minSizeY),
		app.MaxSize(maxSizeX, maxSizeY),
		app.PortraitOrientation.Option(),
		app.NavigationColor(color.NRGBA{A: 0}),
	)

	explorer := expl.NewExplorer(window)

	// font
	fontCollection, err := loadFontCollection()
	if err != nil {
		log.Fatal(err)
	}

	// theme
	theme := material.NewTheme(fontCollection)
	theme.WithPalette(material.Palette{
		Fg:         utils.HexColor(0x000000),
		Bg:         utils.HexColor(0xffffff),
		ContrastBg: utils.HexColor(0x3f51b5),
		ContrastFg: utils.HexColor(0xffffff),
	})

	// main router
	router := router.NewRouter()

	// app instance to give guick access to every package
	app_instance.Window = window
	app_instance.Theme = theme
	app_instance.Router = router
	app_instance.Explorer = explorer

	bottom_bar.LoadInstance()
	node_status_bar.LoadInstance()
	notification_modals.LoadInstance()

	router.Add(app_instance.PAGE_SETTINGS, page_settings.New())
	router.Add(app_instance.PAGE_NODE, page_node.New())
	router.Add(app_instance.PAGE_WALLET, page_wallet.New())
	router.Add(app_instance.PAGE_WALLET_SELECT, page_wallet_select.New())
	router.SetPrimary(app_instance.PAGE_WALLET_SELECT)

	go func() {
		err := runApp()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}
