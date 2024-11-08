package confirm_modal

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/app_instance"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
)

type ConfirmText struct {
	Prompt string
	Yes    string
	No     string
}

type ConfirmModal struct {
	Modal *components.Modal

	confirmText ConfirmText
	buttonYes   *components.Button
	buttonNo    *components.Button

	resChan chan bool
}

var Instance *ConfirmModal

func LoadInstance() {
	modal := components.NewModal(components.ModalStyle{
		CloseOnOutsideClick: true,
		CloseOnInsideClick:  false,
		Direction:           layout.Center,
		Rounded:             components.UniformRounded(unit.Dp(10)),
		Inset:               layout.UniformInset(unit.Dp(10)),
		Animation:           components.NewModalAnimationScaleBounce(),
	})

	buttonYes := components.NewButton(components.ButtonStyle{
		Rounded:   components.UniformRounded(unit.Dp(5)),
		TextSize:  unit.Sp(14),
		Inset:     layout.UniformInset(unit.Dp(10)),
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonYes.Label.Alignment = text.Middle
	buttonYes.Style.Font.Weight = font.Bold

	buttonNo := components.NewButton(components.ButtonStyle{
		Rounded:   components.UniformRounded(unit.Dp(5)),
		TextSize:  unit.Sp(14),
		Inset:     layout.UniformInset(unit.Dp(10)),
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonNo.Label.Alignment = text.Middle
	buttonNo.Style.Font.Weight = font.Bold

	Instance = &ConfirmModal{
		buttonYes: buttonYes,
		buttonNo:  buttonNo,
		Modal:     modal,
	}

	app_instance.Router.AddLayout(router.KeyLayout{
		DrawIndex: 3,
		Layout: func(gtx layout.Context, th *material.Theme) {
			Instance.Layout(gtx, th)
		},
	})
}

func (c *ConfirmModal) Open(confirmText ConfirmText) chan bool {
	c.confirmText = confirmText
	if c.confirmText.Prompt == "" {
		c.confirmText.Prompt = lang.Translate("Are you sure?")
	}

	if c.confirmText.Yes == "" {
		c.confirmText.Yes = lang.Translate("Yes")
	}

	if c.confirmText.No == "" {
		c.confirmText.No = lang.Translate("No")
	}

	c.Modal.SetVisible(true)
	c.resChan = make(chan bool)
	return c.resChan
}

func (c *ConfirmModal) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if c.buttonYes.Clicked() {
		c.resChan <- true
		c.Modal.SetVisible(false)
		close(c.resChan)
	}

	if c.buttonNo.Clicked() {
		c.resChan <- false
		c.Modal.SetVisible(false)
		close(c.resChan)
	}

	var lblSize layout.Dimensions
	c.Modal.Style.Colors = theme.Current.ModalColors
	return c.Modal.Layout(gtx, nil, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Label(th, unit.Sp(18), c.confirmText.Prompt)
					lblSize = label.Layout(gtx)
					return lblSize
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = lblSize.Size.X
					return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							c.buttonNo.Text = c.confirmText.No
							c.buttonNo.Style.Colors = theme.Current.ButtonPrimaryColors
							return c.buttonNo.Layout(gtx, th)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							c.buttonYes.Text = c.confirmText.Yes
							c.buttonYes.Style.Colors = theme.Current.ButtonPrimaryColors
							return c.buttonYes.Layout(gtx, th)
						}),
					)
				}),
			)
		})
	})
}
