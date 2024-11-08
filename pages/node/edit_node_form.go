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
	"github.com/deroproject/derohe/walletapi"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/app_db"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/confirm_modal"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/node_manager"
	"github.com/secretsystems/secret-wallet/prefabs"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/settings"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageEditNodeForm struct {
	isActive bool

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	buttonEdit   *components.Button
	buttonDelete *components.Button
	txtEndpoint  *prefabs.TextField
	txtName      *prefabs.TextField
	nodeConn     app_db.NodeConnection

	list *widget.List
}

var _ router.Page = &PageEditNodeForm{}

func NewPageEditNodeForm() *PageEditNodeForm {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .5, ease.OutCubic),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .5, ease.OutCubic),
	))

	list := new(widget.List)
	list.Axis = layout.Vertical

	saveIcon, _ := widget.NewIcon(icons.ContentSave)
	loadingIcon, _ := widget.NewIcon(icons.NavigationRefresh)
	buttonEdit := components.NewButton(components.ButtonStyle{
		Rounded:     components.UniformRounded(unit.Dp(5)),
		Icon:        saveIcon,
		TextSize:    unit.Sp(14),
		IconGap:     unit.Dp(10),
		Inset:       layout.UniformInset(unit.Dp(10)),
		Animation:   components.NewButtonAnimationDefault(),
		LoadingIcon: loadingIcon,
	})
	buttonEdit.Label.Alignment = text.Middle
	buttonEdit.Style.Font.Weight = font.Bold

	txtName := prefabs.NewTextField()
	txtEndpoint := prefabs.NewTextField()

	deleteIcon, _ := widget.NewIcon(icons.ActionDelete)
	buttonDelete := components.NewButton(components.ButtonStyle{
		Rounded:   components.UniformRounded(unit.Dp(5)),
		Icon:      deleteIcon,
		TextSize:  unit.Sp(14),
		IconGap:   unit.Dp(10),
		Inset:     layout.UniformInset(unit.Dp(10)),
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonDelete.Label.Alignment = text.Middle
	buttonDelete.Style.Font.Weight = font.Bold

	return &PageEditNodeForm{
		animationEnter: animationEnter,
		animationLeave: animationLeave,

		buttonEdit:   buttonEdit,
		buttonDelete: buttonDelete,
		txtName:      txtName,
		txtEndpoint:  txtEndpoint,

		list: list,
	}
}

func (p *PageEditNodeForm) IsActive() bool {
	return p.isActive
}

func (p *PageEditNodeForm) Enter() {
	p.isActive = true
	page_instance.header.Title = func() string { return lang.Translate("Edit Node") }
	p.animationEnter.Start()
	p.animationLeave.Reset()

	p.txtEndpoint.SetValue(p.nodeConn.Endpoint)
	p.txtName.SetValue(p.nodeConn.Name)
}

func (p *PageEditNodeForm) Leave() {
	if page_instance.header.IsHistory(PAGE_EDIT_NODE_FORM) {
		p.animationEnter.Reset()
		p.animationLeave.Start()
	} else {
		p.isActive = false
	}
}

func (p *PageEditNodeForm) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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

	if p.buttonEdit.Clicked() {
		p.submitForm(gtx)
	}

	if p.buttonDelete.Clicked() {
		go func() {
			yesChan := confirm_modal.Instance.Open(confirm_modal.ConfirmText{})

			for yes := range yesChan {
				if yes {
					err := p.removeNode()
					if err != nil {
						notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
						notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
					} else {
						notification_modals.SuccessInstance.SetText(lang.Translate("Success"), lang.Translate("Node deleted"))
						notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
						page_instance.header.GoBack()
					}
				}
			}
		}()
	}

	widgets := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return p.txtName.Layout(gtx, th, lang.Translate("Name"), "Dero NFTs")
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.txtEndpoint.Layout(gtx, th, lang.Translate("Endpoint"), "wss://node.deronfts.com/ws")
		},
		func(gtx layout.Context) layout.Dimensions {
			p.buttonEdit.Text = lang.Translate("SAVE")
			p.buttonEdit.Style.Colors = theme.Current.ButtonPrimaryColors
			return p.buttonEdit.Layout(gtx, th)
		},
		func(gtx layout.Context) layout.Dimensions {
			return prefabs.Divider(gtx, 5)
		},
		func(gtx layout.Context) layout.Dimensions {
			p.buttonDelete.Text = lang.Translate("DELETE NODE")
			p.buttonDelete.Style.Colors = theme.Current.ButtonDangerColors
			return p.buttonDelete.Layout(gtx, th)
		},
	}

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Overlay

	if p.txtName.Input.Clickable.Clicked() {
		p.list.ScrollTo(0)
	}

	if p.txtEndpoint.Input.Clickable.Clicked() {
		p.list.ScrollTo(1)
	}

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(0), Bottom: unit.Dp(20),
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, widgets[index])
	})
}

func (p *PageEditNodeForm) removeNode() error {
	err := app_db.DelNodeConnection(p.nodeConn.ID)
	if err != nil {
		return err
	}

	if node_manager.CurrentNode != nil {
		if node_manager.CurrentNode.Endpoint == p.nodeConn.Endpoint {
			node_manager.CurrentNode = nil
			walletapi.Connected = false

			settings.App.NodeEndpoint = ""
			err := settings.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *PageEditNodeForm) submitForm(gtx layout.Context) {
	p.buttonEdit.SetLoading(true)
	go func() {
		setError := func(err error) {
			p.buttonEdit.SetLoading(false)
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

		node := app_db.NodeConnection{
			ID:          p.nodeConn.ID,
			Name:        txtName.Text(),
			Endpoint:    txtEndpoint.Text(),
			OrderNumber: p.nodeConn.OrderNumber,
		}

		err := app_db.UpdateNodeConnection(node)
		if err != nil {
			setError(err)
			return
		}

		p.buttonEdit.SetLoading(false)
		notification_modals.SuccessInstance.SetText("Success", lang.Translate("Data saved"))
		notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		page_instance.header.GoBack()
	}()
}
