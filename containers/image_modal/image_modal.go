package image_modal

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/app_instance"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type ImageModal struct {
	modal       *components.Modal
	buttonClose *components.Button
	image       *components.Image
	title       string
}

var Instance *ImageModal

func LoadInstance() {
	modal := components.NewModal(components.ModalStyle{
		CloseOnOutsideClick: true,
		CloseOnInsideClick:  false,
		Direction:           layout.N,
		Rounded: components.Rounded{
			SW: unit.Dp(10), SE: unit.Dp(10),
			NW: unit.Dp(10), NE: unit.Dp(10),
		},
		Inset:     layout.UniformInset(unit.Dp(20)),
		Animation: components.NewModalAnimationDown(),
	})

	closeIcon, _ := widget.NewIcon(icons.NavigationCancel)
	buttonClose := components.NewButton(components.ButtonStyle{
		Icon:      closeIcon,
		Animation: components.NewButtonAnimationScale(.95),
	})

	Instance = &ImageModal{
		modal:       modal,
		buttonClose: buttonClose,
	}

	app_instance.Router.AddLayout(router.KeyLayout{
		DrawIndex: 2,
		Layout: func(gtx layout.Context, th *material.Theme) {
			Instance.layout(gtx, th)
		},
	})
}

func (r *ImageModal) Open(title string, imgSrc paint.ImageOp) {
	r.title = title
	r.image = &components.Image{
		Src: imgSrc,
		Fit: components.Contain,
	}
	r.modal.SetVisible(true)
}

func (r *ImageModal) layout(gtx layout.Context, th *material.Theme) {
	if r.buttonClose.Clicked() {
		r.modal.SetVisible(false)
	}

	r.modal.Style.Colors = theme.Current.ModalColors
	r.modal.Layout(gtx, nil, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th, unit.Sp(18), r.title)
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							r.buttonClose.Style.Colors = theme.Current.ModalButtonColors
							return r.buttonClose.Layout(gtx, th)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return r.image.Layout(gtx)
				}),
			)
		})
	})
}
