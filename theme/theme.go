package theme

import (
	"image/color"

	"gioui.org/op/paint"
	"github.com/secretsystems/secret-wallet/assets"
	"github.com/secretsystems/secret-wallet/components"
)

type Theme struct {
	Key            string
	Name           string
	IndicatorColor color.NRGBA

	TextColor            color.NRGBA
	TextMuteColor        color.NRGBA
	DividerColor         color.NRGBA
	BgColor              color.NRGBA
	BgGradientStartColor color.NRGBA
	BgGradientEndColor   color.NRGBA
	HideBalanceBgColor   color.NRGBA

	// Header
	HeaderBackButtonColors components.ButtonColors
	HeaderTopBgColor       color.NRGBA

	// Bottom Bar
	BottomBarBgColor          color.NRGBA
	BottomBarWalletBgColor    color.NRGBA
	BottomBarWalletTextColor  color.NRGBA
	BottomButtonColors        components.ButtonColors
	BottomButtonSelectedColor color.NRGBA

	// Node Status
	NodeStatusBgColor        color.NRGBA
	NodeStatusTextColor      color.NRGBA
	NodeStatusDotGreenColor  color.NRGBA
	NodeStatusDotYellowColor color.NRGBA
	NodeStatusDotRedColor    color.NRGBA

	// Input
	InputColors components.InputColors

	// Button
	ButtonIconPrimaryColors components.ButtonColors
	ButtonPrimaryColors     components.ButtonColors
	ButtonSecondaryColors   components.ButtonColors
	ButtonInvertColors      components.ButtonColors
	ButtonDangerColors      components.ButtonColors

	// Tab Bars
	TabBarsColors components.TabBarsColors

	// Modal
	ModalColors       components.ModalColors
	ModalButtonColors components.ButtonColors

	// Notifications
	NotificationSuccessColors components.NotificationColors
	NotificationErrorColors   components.NotificationColors
	NotificationInfoColors    components.NotificationColors

	// Progress Bar
	ProgressBarColors components.ProgressBarColors

	// List
	ListTextColor        color.NRGBA
	ListBgColor          color.NRGBA
	ListItemHoverBgColor color.NRGBA
	ListScrollBarBgColor color.NRGBA
	ListItemTagBgColor   color.NRGBA
	ListItemTagTextColor color.NRGBA
	//ListItemsColors      components.ListItemsColors

	// Switch
	SwitchColors SwitchColors

	// Images
	ArrowDownArcImage paint.ImageOp
	ArrowUpArcImage   paint.ImageOp
	CoinbaseImage     paint.ImageOp
	TokenImage        paint.ImageOp
	ManageFilesImage  paint.ImageOp
}

type SwitchColors struct {
	Enabled  color.NRGBA
	Disabled color.NRGBA
	Track    color.NRGBA
}

// default to Light theme (avoid nil pointer in FrameEvent before settings.Load() is set)
// settings.Load() will overwrite theme.Current with system pref or settings.json theme value
var Current *Theme = Light

// don't use map[string] the ordering is not guaranteed
var Themes = []*Theme{Light, Dark, Blue}

func Get(key string) *Theme {
	for _, theme := range Themes {
		if theme.Key == key {
			return theme
		}
	}

	return nil
}

func LoadImages() {
	// black
	imgArrowUpArcBlack, _ := assets.GetImage("arrow_up_arc.png")
	opImgArrowUpArcBlack := paint.NewImageOp(imgArrowUpArcBlack)

	imgArrowDownArcBlack, _ := assets.GetImage("arrow_down_arc.png")
	opImgArrowDownArcBlack := paint.NewImageOp(imgArrowDownArcBlack)

	imgCoinbaseBlack, _ := assets.GetImage("coinbase.png")
	opImgCoinbaseBlack := paint.NewImageOp(imgCoinbaseBlack)

	imgTokenBlack, _ := assets.GetImage("token.png")
	opImgTokenBlack := paint.NewImageOp(imgTokenBlack)

	imgManageFilesBlack, _ := assets.GetImage("manage_files.png")
	opImgManageFilesBlack := paint.NewImageOp(imgManageFilesBlack)

	// white
	imgArrowUpArcWhite, _ := assets.GetImage("arrow_up_arc_white.png")
	opImgArrowUpArcWhite := paint.NewImageOp(imgArrowUpArcWhite)

	imgArrowDownArcWhite, _ := assets.GetImage("arrow_down_arc_white.png")
	opImgArrowDownArcWhite := paint.NewImageOp(imgArrowDownArcWhite)

	imgCoinbaseWhite, _ := assets.GetImage("coinbase_white.png")
	opImgCoinbaseWhite := paint.NewImageOp(imgCoinbaseWhite)

	imgTokenWhite, _ := assets.GetImage("token_white.png")
	opImgTokenWhite := paint.NewImageOp(imgTokenWhite)

	imgManageFilesWhite, _ := assets.GetImage("manage_files_white.png")
	opImgManageFilesWhite := paint.NewImageOp(imgManageFilesWhite)

	// light theme
	Light.ArrowUpArcImage = opImgArrowUpArcBlack
	Light.ArrowDownArcImage = opImgArrowDownArcBlack
	Light.CoinbaseImage = opImgCoinbaseBlack
	Light.ManageFilesImage = opImgManageFilesBlack
	Light.TokenImage = opImgTokenBlack

	// dark theme
	Dark.ArrowUpArcImage = opImgArrowUpArcWhite
	Dark.ArrowDownArcImage = opImgArrowDownArcWhite
	Dark.CoinbaseImage = opImgCoinbaseWhite
	Dark.ManageFilesImage = opImgManageFilesWhite
	Dark.TokenImage = opImgTokenWhite

	//blue
	Blue.ArrowUpArcImage = opImgArrowUpArcWhite
	Blue.ArrowDownArcImage = opImgArrowDownArcWhite
	Blue.CoinbaseImage = opImgCoinbaseWhite
	Blue.ManageFilesImage = opImgManageFilesWhite
	Blue.TokenImage = opImgTokenWhite
}
