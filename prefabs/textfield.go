package prefabs

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/theme"
)

type TextField struct {
	Input *Input
}

func NewTextField() *TextField {
	input := NewInput()

	return &TextField{
		Input: input,
	}
}

func NewNumberTextField() *TextField {
	input := NewNumberInput()
	return &TextField{
		Input: input,
	}
}

func NewPasswordTextField() *TextField {
	input := NewPasswordInput()

	return &TextField{
		Input: input,
	}
}

func (t *TextField) Value() string {
	return t.Input.Value()
}

func (t *TextField) SetValue(text string) {
	t.Input.SetValue(text)
}

func (t *TextField) Editor() *widget.Editor {
	return t.Input.Editor
}

func (t *TextField) Layout(gtx layout.Context, th *material.Theme, title string, hint string) layout.Dimensions {
	t.Input.Colors = theme.Current.InputColors
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			textSize := th.TextSize + unit.Sp(2)
			lbl := material.Label(th, textSize, title)
			lbl.Color = t.Input.Colors.TextColor
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.Input.Layout(gtx, th, hint)
		}),
	)
}
