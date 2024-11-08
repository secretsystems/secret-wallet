package prefabs

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

type ListItemSelect struct {
	visible        bool
	animationEnter *animation.Animation
	animationLeave *animation.Animation
}

func NewListItemSelect() *ListItemSelect {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .15, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .15, ease.Linear),
	))

	return &ListItemSelect{
		animationEnter: animationEnter,
		animationLeave: animationLeave,
	}
}

func (n *ListItemSelect) Toggle() {
	n.SetVisible(!n.visible)
}

func (n *ListItemSelect) SetVisible(visible bool) {
	if visible {
		n.visible = true
		n.animationEnter.Start()
		n.animationLeave.Reset()
	} else {
		n.animationEnter.Reset()
		n.animationLeave.Start()
	}
}

func (n *ListItemSelect) Layout(gtx layout.Context, th *material.Theme, firstButton *components.Button, secondButton *components.Button) layout.Dimensions {
	if !n.visible {
		return layout.Dimensions{}
	}

	{
		state := n.animationEnter.Update(gtx)
		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}
	}

	{
		state := n.animationLeave.Update(gtx)

		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}

		if state.Finished {
			n.visible = false
			op.InvalidateOp{}.Add(gtx.Ops)
		}
	}

	return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return firstButton.Layout(gtx, th)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return secondButton.Layout(gtx, th)
			}),
		)
	})
}

type ListItemMenuItem struct {
	Key   string
	Icon  *widget.Icon
	Title string
}

func (t ListItemMenuItem) Layout(gtx layout.Context, index int, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = gtx.Dp(45)
			gtx.Constraints.Max.Y = gtx.Dp(30)
			return t.Icon.Layout(gtx, th.Fg)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(20), lang.Translate(t.Title))
			return lbl.Layout(gtx)
		}),
	)
}
