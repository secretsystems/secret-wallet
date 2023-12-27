package page_settings

import (
	"fmt"
	"net"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi/rpcserver"
	"github.com/g45t345rt/g45w/animation"
	"github.com/g45t345rt/g45w/components"
	"github.com/g45t345rt/g45w/containers/notification_modals"
	"github.com/g45t345rt/g45w/lang"
	"github.com/g45t345rt/g45w/prefabs"
	"github.com/g45t345rt/g45w/router"
	"github.com/g45t345rt/g45w/theme"
	"github.com/g45t345rt/g45w/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageRpc struct {
	isActive bool

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	list      *widget.List
	rpcServer *RpcServer
}

type RpcServer struct {
	user      string
	pass      string
	buttonOn  *components.Button
	buttonOff *components.Button
	txtUser   *prefabs.TextField
	txtPass   *prefabs.TextField
	server    *rpcserver.RPCServer
}

func NewRPCServer() *RpcServer {

	addIcon, _ := widget.NewIcon(icons.ContentAdd)
	loadingIcon, _ := widget.NewIcon(icons.NavigationRefresh)

	buttonOn := components.NewButton(components.ButtonStyle{
		Rounded:     components.UniformRounded(unit.Dp(5)),
		Icon:        addIcon,
		TextSize:    unit.Sp(14),
		IconGap:     unit.Dp(10),
		Inset:       layout.UniformInset(unit.Dp(10)),
		Animation:   components.NewButtonAnimationDefault(),
		LoadingIcon: loadingIcon,
	})
	buttonOn.Label.Alignment = text.Middle
	buttonOn.Style.Font.Weight = font.Bold

	buttonOff := components.NewButton(components.ButtonStyle{
		Rounded:     components.UniformRounded(unit.Dp(5)),
		Icon:        addIcon,
		TextSize:    unit.Sp(14),
		IconGap:     unit.Dp(10),
		Inset:       layout.UniformInset(unit.Dp(10)),
		Animation:   components.NewButtonAnimationDefault(),
		LoadingIcon: loadingIcon,
	})
	buttonOff.Label.Alignment = text.Middle
	buttonOff.Style.Font.Weight = font.Bold

	txtUser := prefabs.NewTextField()
	txtPass := prefabs.NewTextField()

	item := &RpcServer{
		user:      "",
		pass:      "",
		buttonOn:  buttonOn,
		buttonOff: buttonOff,
		txtUser:   txtUser,
		txtPass:   txtPass,
		server:    nil,
	}
	return item
}

var _ router.Page = &PageRpc{}

func NewPageRpc() *PageRpc {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .25, ease.Linear),
	))

	list := new(widget.List)
	list.Axis = layout.Vertical

	return &PageRpc{
		animationEnter: animationEnter,
		animationLeave: animationLeave,

		rpcServer: NewRPCServer(),

		list: list,
	}
}

func (p *PageRpc) IsActive() bool {
	return p.isActive
}

func (p *PageRpc) Enter() {
	p.isActive = true
	page_instance.header.Title = func() string { return lang.Translate("About RPC") }

	if !page_instance.header.IsHistory(PAGE_APP_INFO) {
		p.animationEnter.Start()
		p.animationLeave.Reset()
	}
}

func (p *PageRpc) Leave() {
	p.animationEnter.Reset()
	p.animationLeave.Start()
}

func (p *PageRpc) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	{
		state := p.animationEnter.Update(gtx)
		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}
	}

	{
		state := p.animationLeave.Update(gtx)
		if state.Finished {
			p.isActive = false
			op.InvalidateOp{}.Add(gtx.Ops)
		}

		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}
	}

	if p.rpcServer.buttonOn.Clicked() {
		p.turnOn(gtx)
	}
	if p.rpcServer.buttonOff.Clicked() {
		p.turnOff(gtx)
	}

	var widgets []layout.Widget

	widgets = append(
		widgets,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
		}, func(gtx layout.Context) layout.Dimensions {
			address, _ := getLocalIP()
			lbl := material.Label(th, unit.Sp(16), lang.Translate("Set RPC Username Password \nDefault IP:Port = "+address+"10107"))
			lbl.Color = theme.Current.TextMuteColor
			return lbl.Layout(gtx)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.rpcServer.txtUser.Layout(gtx, th, lang.Translate("Username"), "RPC username")
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.rpcServer.txtPass.Layout(gtx, th, lang.Translate("Password"), "RPC password")
		},
		func(gtx layout.Context) layout.Dimensions {
			p.rpcServer.buttonOn.Text = lang.Translate("Turn RPC on")
			p.rpcServer.buttonOn.Style.Colors = theme.Current.ButtonPrimaryColors
			return p.rpcServer.buttonOn.Layout(gtx, th)
		}, func(gtx layout.Context) layout.Dimensions {
			p.rpcServer.buttonOff.Text = lang.Translate("Turn RPC off")
			p.rpcServer.buttonOff.Style.Colors = theme.Current.ButtonPrimaryColors
			return p.rpcServer.buttonOff.Layout(gtx, th)
		})

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Overlay

	if p.rpcServer.txtUser.Input.Clickable.Clicked() {
		p.list.ScrollTo(0)
	}

	if p.rpcServer.txtPass.Input.Clickable.Clicked() {
		p.list.ScrollTo(0)
	}

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(0), Bottom: unit.Dp(20),
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, widgets[index])
	})
}

func (p *PageRpc) turnOn(gtx layout.Context) {
	var err error

	go func() {

		setError := func(err error) {
			p.rpcServer.buttonOn.SetLoading(false)
			notification_modals.ErrorInstance.SetText("Error", err.Error())
			notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		}

		if wallet_manager.OpenedWallet == nil {
			setError(fmt.Errorf("No opened wallet"))
			return
		}

		txtUser := p.rpcServer.txtUser.Editor()
		txtPass := p.rpcServer.txtPass.Editor()

		if txtUser.Text() == "" {
			setError(fmt.Errorf("enter user"))
			return
		}

		if txtPass.Text() == "" {
			setError(fmt.Errorf("enter pass"))
			return
		}
		globals.Arguments["--rpc-server"] = true
		globals.Arguments["--rpc-login"] = p.rpcServer.txtUser.Value() + ":" + p.rpcServer.txtPass.Value()
		globals.Arguments["--rpc-bind"] = "0.0.0.0:10107"

		p.rpcServer.server, err = rpcserver.RPCServer_Start(wallet_manager.OpenedWallet.Memory, "secret-wallet")
		wallet_manager.OpenedWallet.Server = p.rpcServer.server
		if err != nil {
			p.rpcServer.server = nil
		}
		//todo
		//please do a quick check to see if you are connected :(
		// _, err := rpcserver.GetAddress()
		// if err != nil {
		// 	setError(err)
		// 	return
		// }

		p.rpcServer.buttonOn.SetLoading(false)
		notification_modals.SuccessInstance.SetText(lang.Translate("Success"), "RPC turned on")
		notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		page_instance.header.GoBack()
	}()
}

func (p *PageRpc) turnOff(gtx layout.Context) {
	setError := func(err error) {
		p.rpcServer.buttonOff.SetLoading(false)
		notification_modals.ErrorInstance.SetText("Error", err.Error())
		notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
	}

	if wallet_manager.OpenedWallet == nil {
		setError(fmt.Errorf("No opened wallet"))
		return
	}
	p.rpcServer.buttonOff.SetLoading(true)
	go func() {
		globals.Arguments["--rpc-server"] = false
		p.rpcServer.server.RPCServer_Stop()

		p.rpcServer.buttonOff.SetLoading(false)
		notification_modals.SuccessInstance.SetText(lang.Translate("Success"), "RPC turned off")
		notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		page_instance.header.GoBack()
	}()
}

func getLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				if ipNet.IP.String()[:3] == "192" && ipNet.IP.String()[4:7] == "168" {
					return ipNet.IP.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("Local IP address not found")
}
