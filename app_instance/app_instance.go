package app_instance

import (
	"image/color"

	"gioui.org/app"
	"gioui.org/unit"
	"gioui.org/x/explorer"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/settings"
)

var Window *app.Window
var Router *router.Router
var Explorer *explorer.Explorer

func Load() {
	minSizeX := unit.Dp(375)
	minSizeY := unit.Dp(600)
	maxSizeX := unit.Dp(500)
	maxSizeY := unit.Dp(1000)

	Window = app.NewWindow(
		app.Title(settings.Name),
		app.MinSize(minSizeX, minSizeY),
		app.Size(minSizeX, minSizeY),
		app.MaxSize(maxSizeX, maxSizeY),
		app.PortraitOrientation.Option(),
		app.NavigationColor(color.NRGBA{A: 0}),
	)

	Explorer = explorer.NewExplorer(Window)
	Router = router.NewRouter()
}
