package page_wallet

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/g45t345rt/g45w/app_instance"
	"github.com/g45t345rt/g45w/router"
	"github.com/g45t345rt/g45w/ui/animation"
	"github.com/g45t345rt/g45w/ui/components"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageSendForm struct {
	isActive bool

	txtAmount     *components.TextField
	txtWalletAddr *components.TextField
	txtComment    *components.TextField
	txtDstPort    *components.TextField
	buttonBuildTx *components.Button

	ringSizeSelector *RingSizeSelector

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	listStyle material.ListStyle
}

var _ router.Container = &PageSendForm{}

func NewPageSendForm() *PageSendForm {
	th := app_instance.Current.Theme
	buildIcon, _ := widget.NewIcon(icons.HardwareMemory)
	buttonBuildTx := components.NewButton(components.ButtonStyle{
		Rounded:         unit.Dp(5),
		Text:            "BUILD TRANSACTION",
		Icon:            buildIcon,
		TextColor:       color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		BackgroundColor: color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		TextSize:        unit.Sp(14),
		IconGap:         unit.Dp(10),
		Inset:           layout.UniformInset(unit.Dp(10)),
		Animation:       components.NewButtonAnimationDefault(),
	})
	buttonBuildTx.Label.Alignment = text.Middle
	buttonBuildTx.Style.Font.Weight = font.Bold

	txtAmount := components.NewTextField(th, "Amount", "")
	txtWalletAddr := components.NewTextField(th, "Wallet Addr / Name", "")
	txtComment := components.NewTextField(th, "Comment", "The comment is natively encrypted.")
	txtComment.EditorStyle.Editor.SingleLine = false
	txtComment.EditorStyle.Editor.Submit = false
	txtDstPort := components.NewTextField(th, "Destination Port", "")

	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .25, ease.Linear),
	))

	list := new(widget.List)
	list.Axis = layout.Vertical
	listStyle := material.List(th, list)
	listStyle.AnchorStrategy = material.Overlay

	return &PageSendForm{
		txtAmount:        txtAmount,
		txtWalletAddr:    txtWalletAddr,
		txtComment:       txtComment,
		txtDstPort:       txtDstPort,
		buttonBuildTx:    buttonBuildTx,
		ringSizeSelector: NewRingSizeSelector("16"),
		animationEnter:   animationEnter,
		animationLeave:   animationLeave,
		listStyle:        listStyle,
	}
}

func (p *PageSendForm) IsActive() bool {
	return p.isActive
}

func (p *PageSendForm) Enter() {
	p.isActive = true
	p.animationEnter.Start()
	p.animationLeave.Reset()
}

func (p *PageSendForm) Leave() {
	p.animationLeave.Start()
	p.animationEnter.Reset()
}

func (p *PageSendForm) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.buttonBuildTx.Clickable.Clicked() {

	}

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

	widgets := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					labelTitle := material.Label(th, unit.Sp(18), "Send DERO")
					return labelTitle.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					labelTokenId := material.Label(th, unit.Sp(14), "00000...00000")
					labelTokenId.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 150}
					return labelTokenId.Layout(gtx)
				}),
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.txtAmount.Layout(gtx, th)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.txtWalletAddr.Layout(gtx, th)
		},
		func(gtx layout.Context) layout.Dimensions {
			p.txtComment.EditorMinY = gtx.Dp(75)
			return p.txtComment.Layout(gtx, th)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.txtDstPort.Layout(gtx, th)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.buttonBuildTx.Layout(gtx, th)
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
		},
	}

	return p.listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(0), Bottom: unit.Dp(20),
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, widgets[index])
	})
}

type RingSizeSelector struct {
	btn *components.Button
}

func NewRingSizeSelector(defaultSize string) *RingSizeSelector {
	btn := components.NewButton(components.ButtonStyle{
		Rounded:         unit.Dp(5),
		Text:            defaultSize,
		TextColor:       color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		BackgroundColor: color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		TextSize:        unit.Sp(14),
		Inset:           layout.UniformInset(unit.Dp(10)),
	})
	btn.Label.Alignment = text.Middle
	btn.Style.Font.Weight = font.Bold

	//sizes := []string{"2", "4", "8", "16", "32", "64", "128", "256"}

	return &RingSizeSelector{
		btn: btn,
	}
}

func (r *RingSizeSelector) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return r.btn.Layout(gtx, th)
}
