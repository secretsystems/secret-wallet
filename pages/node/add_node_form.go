package page_node

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/app_db"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/prefabs"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageAddNodeForm struct {
	isActive bool

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	buttonAdd   *components.Button
	txtEndpoint *prefabs.TextField
	txtName     *prefabs.TextField

	list *widget.List
}

var _ router.Page = &PageAddNodeForm{}

func NewPageAddNodeForm() *PageAddNodeForm {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .5, ease.OutCubic),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .5, ease.OutCubic),
	))

	list := new(widget.List)
	list.Axis = layout.Vertical

	addIcon, _ := widget.NewIcon(icons.ContentAdd)
	loadingIcon, _ := widget.NewIcon(icons.NavigationRefresh)
	buttonAdd := components.NewButton(components.ButtonStyle{
		Rounded:     components.UniformRounded(unit.Dp(5)),
		Icon:        addIcon,
		TextSize:    unit.Sp(14),
		IconGap:     unit.Dp(10),
		Inset:       layout.UniformInset(unit.Dp(10)),
		Animation:   components.NewButtonAnimationDefault(),
		LoadingIcon: loadingIcon,
	})
	buttonAdd.Label.Alignment = text.Middle
	buttonAdd.Style.Font.Weight = font.Bold

	txtName := prefabs.NewTextField()
	txtEndpoint := prefabs.NewTextField()

	return &PageAddNodeForm{
		animationEnter: animationEnter,
		animationLeave: animationLeave,

		buttonAdd:   buttonAdd,
		txtName:     txtName,
		txtEndpoint: txtEndpoint,

		list: list,
	}
}

func (p *PageAddNodeForm) IsActive() bool {
	return p.isActive
}

func (p *PageAddNodeForm) Enter() {
	p.isActive = true
	page_instance.header.Title = func() string { return lang.Translate("Add Node") }
	p.animationEnter.Start()
	p.animationLeave.Reset()
}

func (p *PageAddNodeForm) Leave() {
	if page_instance.header.IsHistory(PAGE_ADD_NODE_FORM) {
		p.animationEnter.Reset()
		p.animationLeave.Start()
	} else {
		p.isActive = false
	}
}

func (p *PageAddNodeForm) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	{
		state := p.animationEnter.Update(gtx)
		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}
	}

	{
		state := p.animationLeave.Update(gtx)
		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}

		if state.Finished {
			p.isActive = false
			op.InvalidateOp{}.Add(gtx.Ops)
		}
	}

	if p.buttonAdd.Clicked() {
		p.submitForm(gtx)
	}

	widgets := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(16), lang.Translate("Here, you can add your own remote node. The endpoint connection must be a WebSocket connection, starting with ws:// or wss:// for TLS connection."))
			lbl.Color = theme.Current.TextMuteColor
			return lbl.Layout(gtx)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.txtName.Layout(gtx, th, lang.Translate("Name"), "Local")
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.txtEndpoint.Layout(gtx, th, lang.Translate("Endpoint"), "ws://127.0.0.1:10102/ws")
		},
		func(gtx layout.Context) layout.Dimensions {
			p.buttonAdd.Text = lang.Translate("ADD NODE")
			p.buttonAdd.Style.Colors = theme.Current.ButtonPrimaryColors
			return p.buttonAdd.Layout(gtx, th)
		},
	}

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Overlay

	if p.txtName.Input.Clickable.Clicked() {
		p.list.ScrollTo(0)
	}

	if p.txtEndpoint.Input.Clickable.Clicked() {
		p.list.ScrollTo(0)
	}

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(0), Bottom: unit.Dp(20),
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, widgets[index])
	})
}

func (p *PageAddNodeForm) submitForm(gtx layout.Context) {
	p.buttonAdd.SetLoading(true)
	go func() {
		setError := func(err error) {
			p.buttonAdd.SetLoading(false)
			notification_modals.ErrorInstance.SetText("Error", err.Error())
			notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		}

		txtName := p.txtName.Editor()
		txtEndpoint := p.txtEndpoint.Editor()

		if txtName.Text() == "" {
			setError(fmt.Errorf("enter name"))
			return
		}

		if txtEndpoint.Text() == "" {
			setError(fmt.Errorf("enter endpoint"))
			return
		}

		err := app_db.InsertNodeConnection(app_db.NodeConnection{
			Name:     txtName.Text(),
			Endpoint: txtEndpoint.Text(),
		})
		if err != nil {
			setError(err)
			return
		}

		p.buttonAdd.SetLoading(false)
		notification_modals.SuccessInstance.SetText(lang.Translate("Success"), "new noded added")
		notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		page_instance.header.GoBack()
	}()
}
