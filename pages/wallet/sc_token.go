package page_wallet

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"strconv"

	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/deroproject/derohe/rpc"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/app_icons"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/build_tx_modal"
	"github.com/secretsystems/secret-wallet/containers/confirm_modal"
	"github.com/secretsystems/secret-wallet/containers/image_modal"
	"github.com/secretsystems/secret-wallet/containers/listselect_modal"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/containers/prompt_modal"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/prefabs"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/sc"
	"github.com/secretsystems/secret-wallet/settings"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/secretsystems/secret-wallet/utils"
	"github.com/secretsystems/secret-wallet/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageSCToken struct {
	isActive bool

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	buttonOpenMenu      *components.Button
	sendReceiveButtons  *SendReceiveButtons
	tabBars             *components.TabBars
	txBar               *TxBar
	getEntriesParams    wallet_manager.GetEntriesParams
	txItems             []*TxListItem
	tokenInfo           *TokenInfoList
	balanceContainer    *BalanceContainer
	g45DisplayContainer *G45DisplayContainer
	buttonCopySCID      *components.Button

	token      *wallet_manager.Token
	scIdEditor *widget.Editor

	list *widget.List
}

var _ router.Page = &PageSCToken{}

func NewPageSCToken() *PageSCToken {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .25, ease.Linear),
	))

	addIcon, _ := widget.NewIcon(icons.NavigationMenu)
	buttonOpenMenu := components.NewButton(components.ButtonStyle{
		Icon:      addIcon,
		Animation: components.NewButtonAnimationScale(.98),
	})

	list := new(widget.List)
	list.Axis = layout.Vertical

	scIdEditor := new(widget.Editor)
	scIdEditor.WrapPolicy = text.WrapGraphemes
	scIdEditor.ReadOnly = true

	sendReceiveButtons := NewSendReceiveButtons()

	tabBarsItems := []*components.TabBarsItem{
		components.NewTabBarItem("txs"),
		components.NewTabBarItem("info"),
	}

	tabBars := components.NewTabBars("txs", tabBarsItems)
	txBar := NewTxBar()

	balanceContainer := NewBalanceContainer()
	g45DisplayContainer := NewG45DisplayContainer()

	copyIcon, _ := widget.NewIcon(icons.ContentContentCopy)
	buttonCopySCID := components.NewButton(components.ButtonStyle{
		Icon: copyIcon,
	})

	return &PageSCToken{
		animationEnter: animationEnter,
		animationLeave: animationLeave,

		buttonOpenMenu:      buttonOpenMenu,
		scIdEditor:          scIdEditor,
		sendReceiveButtons:  sendReceiveButtons,
		tabBars:             tabBars,
		txBar:               txBar,
		balanceContainer:    balanceContainer,
		g45DisplayContainer: g45DisplayContainer,
		buttonCopySCID:      buttonCopySCID,

		list: list,
	}
}

func (p *PageSCToken) IsActive() bool {
	return p.isActive
}

func (p *PageSCToken) Enter() {
	p.isActive = true

	wallet := wallet_manager.OpenedWallet
	wallet.Memory.TokenAdd(p.token.GetHash()) // we don't check error because the only possible error is if the token was already added

	p.tokenInfo = NewTokenInfoList(p.token)

	page_instance.header.Title = func() string { return p.token.Name }
	page_instance.header.Subtitle = func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		if p.buttonCopySCID.Clicked() {
			clipboard.WriteOp{
				Text: p.token.SCID,
			}.Add(gtx.Ops)
			notification_modals.InfoInstance.SetText(lang.Translate("Clipboard"), lang.Translate("SCID copied to clipboard"))
			notification_modals.InfoInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
		}

		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				scId := utils.ReduceTxId(p.token.SCID)
				if p.token.Symbol.Valid {
					scId = fmt.Sprintf("%s (%s)", scId, p.token.Symbol.String)
				}

				lbl := material.Label(th, unit.Sp(16), scId)
				lbl.Color = theme.Current.TextMuteColor
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(5)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = gtx.Dp(18)
				gtx.Constraints.Max.Y = gtx.Dp(18)
				p.buttonCopySCID.Style.Colors = theme.Current.ModalButtonColors
				return p.buttonCopySCID.Layout(gtx, th)
			}),
		)
	}
	page_instance.header.ButtonRight = p.buttonOpenMenu
	p.scIdEditor.SetText(p.token.SCID)

	if !page_instance.header.IsHistory(PAGE_SC_TOKEN) {
		p.animationEnter.Start()
		p.animationLeave.Reset()
	}
}

func (p *PageSCToken) Leave() {
	p.animationLeave.Start()
	p.animationEnter.Reset()
}

func (p *PageSCToken) LoadTxs() {
	wallet := wallet_manager.OpenedWallet
	hash := p.token.GetHash()
	entries := wallet.GetEntries(&hash, p.getEntriesParams)

	txItems := []*TxListItem{}

	for _, entry := range entries {
		txItems = append(txItems, NewTxListItem(entry, int(p.token.Decimals)))
	}

	p.txItems = txItems
	p.txBar.txCount = len(entries)
}

func (p *PageSCToken) SetToken(token *wallet_manager.Token) {
	p.token = token
	p.token.RefreshImageOp()
	p.balanceContainer.SetToken(p.token)
	p.g45DisplayContainer.SetToken(p.token)
	p.g45DisplayContainer.Load()
	p.LoadTxs()
}

func (p *PageSCToken) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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

	if p.buttonOpenMenu.Clickable.Clicked() {
		go func() {
			showIcon, _ := widget.NewIcon(icons.ActionVisibility)
			hideIcon, _ := widget.NewIcon(icons.ActionVisibilityOff)
			refreshIcon, _ := widget.NewIcon(icons.NavigationRefresh)
			addFavIcon, _ := widget.NewIcon(icons.ToggleStarBorder)
			delFavIcon, _ := widget.NewIcon(icons.ToggleStar)
			//editIcon, _ := widget.NewIcon(icons.ActionInput)
			deleteIcon, _ := widget.NewIcon(icons.ActionDelete)
			ethereumIcon, _ := widget.NewIcon(app_icons.Ethereum)

			var items []*listselect_modal.SelectListItem
			token := page_instance.pageSCToken.token

			isFav := false
			if token != nil && token.IsFavorite.Valid {
				isFav = token.IsFavorite.Bool
			}

			standardType := sc.UNKNOWN_TYPE
			if token != nil {
				standardType = token.StandardType
			}

			if standardType == sc.G45_NFT_TYPE {
				items = append(items, listselect_modal.NewSelectListItem("g45_display_nft",
					listselect_modal.NewItemText(showIcon, lang.Translate("Display NFT")).Layout,
				))

				items = append(items, listselect_modal.NewSelectListItem("g45_retrieve_nft",
					listselect_modal.NewItemText(hideIcon, lang.Translate("Retrieve NFT")).Layout,
				))
			}

			if standardType == sc.G45_AT_TYPE || standardType == sc.G45_FAT_TYPE {
				items = append(items, listselect_modal.NewSelectListItem("g45_display_token",
					listselect_modal.NewItemText(showIcon, lang.Translate("Display tokens")).Layout,
				))

				items = append(items, listselect_modal.NewSelectListItem("g45_retrieve_token",
					listselect_modal.NewItemText(hideIcon, lang.Translate("Retrieve tokens")).Layout,
				))
			}

			if standardType == sc.DEX_SC_TYPE {
				items = append(items, listselect_modal.NewSelectListItem("dex_sc_bridge_in",
					listselect_modal.NewItemText(ethereumIcon, lang.Translate("Bridge in")).Layout,
				))

				items = append(items, listselect_modal.NewSelectListItem("dex_sc_bridge_out",
					listselect_modal.NewItemText(ethereumIcon, lang.Translate("Bridge out")).Layout,
				))
			}

			items = append(items, listselect_modal.NewSelectListItem("refresh_cache",
				listselect_modal.NewItemText(refreshIcon, lang.Translate("Refresh cache")).Layout,
			))

			if !isFav {
				items = append(items, listselect_modal.NewSelectListItem("add_favorite",
					listselect_modal.NewItemText(addFavIcon, lang.Translate("Add to favorites")).Layout,
				))
			}

			if isFav {
				items = append(items, listselect_modal.NewSelectListItem("remove_favorite",
					listselect_modal.NewItemText(delFavIcon, lang.Translate("Remove from favorites")).Layout,
				))
			}

			/*
				items = append(items, listselect_modal.NewSelectListItem("edit_token",
					listselect_modal.NewItemText(editIcon, lang.Translate("Edit token")).Layout,
				))
			*/

			items = append(items, listselect_modal.NewSelectListItem("remove_token",
				listselect_modal.NewItemText(deleteIcon, lang.Translate("Remove token")).Layout,
			))

			keyChan := listselect_modal.Instance.Open(items)

			for sKey := range keyChan {
				wallet := wallet_manager.OpenedWallet
				var err error
				var successMsg = ""

				switch sKey {
				case "refresh_cache":
					wallet.ResetBalanceResult(p.token.SCID)
					successMsg = lang.Translate("Cache refreshed.")
				case "add_favorite":
					p.token.IsFavorite = sql.NullBool{Bool: true, Valid: true}
					err = wallet.UpdateToken(*p.token)
					successMsg = lang.Translate("Token added to favorites.")
				case "remove_favorite":
					p.token.IsFavorite = sql.NullBool{Bool: false, Valid: true}
					err = wallet.UpdateToken(*p.token)
					successMsg = lang.Translate("Token removed from favorites.")
				case "remove_token":
					yesChan := confirm_modal.Instance.Open(confirm_modal.ConfirmText{})

					for yes := range yesChan {
						if yes {
							wallet := wallet_manager.OpenedWallet
							err := wallet.DelToken(p.token.ID)

							if err != nil {
								notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
								notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
							} else {
								page_instance.header.GoBack()
								notification_modals.SuccessInstance.SetText(lang.Translate("Success"), lang.Translate("Token removed."))
								notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
							}
						}
					}
				case "g45_display_nft":
					scId := p.token.GetHash()
					build_tx_modal.Instance.OpenWithRandomAddr(scId, func(randomAddr string, open func(txPayload build_tx_modal.TxPayload)) {
						open(build_tx_modal.TxPayload{
							Transfers: []rpc.Transfer{
								{SCID: scId, Destination: randomAddr, Burn: uint64(1)},
							},
							Ringsize: 2,
							SCArgs: rpc.Arguments{
								{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
								{Name: rpc.SCID, DataType: rpc.DataHash, Value: scId},
								{Name: "entrypoint", DataType: rpc.DataString, Value: "DisplayNFT"},
							},
							TokensInfo: []*wallet_manager.Token{p.token},
						})
					})
				case "g45_retrieve_nft":
					scId := p.token.GetHash()
					build_tx_modal.Instance.Open(build_tx_modal.TxPayload{
						Ringsize: 2,
						SCArgs: rpc.Arguments{
							{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
							{Name: rpc.SCID, DataType: rpc.DataHash, Value: scId},
							{Name: "entrypoint", DataType: rpc.DataString, Value: "RetrieveNFT"},
						},
					})
				case "g45_display_token":
					txtChan := prompt_modal.Instance.Open("", lang.Translate("Enter amount"), key.HintNumeric)
					for txt := range txtChan {
						amount := utils.ShiftNumber{Decimals: int(p.token.Decimals)}
						err := amount.Parse(txt)
						if err != nil {
							return
						}

						scId := p.token.GetHash()
						build_tx_modal.Instance.OpenWithRandomAddr(scId, func(randomAddr string, open func(txPayload build_tx_modal.TxPayload)) {
							open(build_tx_modal.TxPayload{
								Transfers: []rpc.Transfer{
									{SCID: scId, Destination: randomAddr, Burn: amount.Number},
								},
								Ringsize: 2,
								SCArgs: rpc.Arguments{
									{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
									{Name: rpc.SCID, DataType: rpc.DataHash, Value: scId},
									{Name: "entrypoint", DataType: rpc.DataString, Value: "DisplayToken"},
								},
								TokensInfo: []*wallet_manager.Token{p.token},
							})
						})
					}
				case "g45_retrieve_token":
					txtChan := prompt_modal.Instance.Open("", lang.Translate("Enter amount"), key.HintNumeric)
					for txt := range txtChan {
						amount := utils.ShiftNumber{Decimals: int(p.token.Decimals)}
						err := amount.Parse(txt)
						if err != nil {
							return
						}

						scId := p.token.GetHash()
						build_tx_modal.Instance.OpenWithRandomAddr(scId, func(randomAddr string, open func(txPayload build_tx_modal.TxPayload)) {
							open(build_tx_modal.TxPayload{
								Ringsize: 2,
								SCArgs: rpc.Arguments{
									{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
									{Name: rpc.SCID, DataType: rpc.DataHash, Value: scId},
									{Name: "entrypoint", DataType: rpc.DataString, Value: "RetrieveToken"},
									{Name: "amount", DataType: rpc.DataUint64, Value: amount.Number},
								},
							})
						})
					}
					// case "dex_sc_bridge_in":
					// 	page_instance.pageDEXSCBridgeIn.SetToken(p.token)
					// 	page_instance.pageRouter.SetCurrent(PAGE_DEX_SC_BRIDGE_IN)
					// 	page_instance.header.AddHistory(PAGE_DEX_SC_BRIDGE_IN)
					// case "dex_sc_bridge_out":
					// 	page_instance.pageDEXSCBridgeOut.SetToken(p.token)
					// 	page_instance.pageRouter.SetCurrent(PAGE_DEX_SC_BRIDGE_OUT)
					// 	page_instance.header.AddHistory(PAGE_DEX_SC_BRIDGE_OUT)
				}

				if err != nil {
					notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
					notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
				} else if successMsg != "" {
					notification_modals.SuccessInstance.SetText(lang.Translate("Success"), successMsg)
					notification_modals.SuccessInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
				}
			}
		}()
	}

	{
		changed, tab := p.txBar.Changed()
		if changed {
			switch tab {
			case "all":
				p.getEntriesParams = wallet_manager.GetEntriesParams{}
			case "in":
				p.getEntriesParams = wallet_manager.GetEntriesParams{
					In: sql.NullBool{Bool: true, Valid: true},
				}
			case "out":
				p.getEntriesParams = wallet_manager.GetEntriesParams{
					Out: sql.NullBool{Bool: true, Valid: true},
				}
			case "coinbase":
				p.getEntriesParams = wallet_manager.GetEntriesParams{
					Coinbase: sql.NullBool{Bool: true, Valid: true},
				}
			}

			p.LoadTxs()
		}
	}

	if p.sendReceiveButtons.ButtonSend.Clicked() {
		page_instance.pageSendForm.SetToken(p.token)
		page_instance.pageSendForm.ClearForm()
		page_instance.pageRouter.SetCurrent(PAGE_SEND_FORM)
		page_instance.header.AddHistory(PAGE_SEND_FORM)
	}

	if p.sendReceiveButtons.ButtonReceive.Clicked() {
		page_instance.pageRouter.SetCurrent(PAGE_RECEIVE_FORM)
		page_instance.header.AddHistory(PAGE_RECEIVE_FORM)
	}

	widgets := []layout.Widget{}

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Overlay

	switch p.token.StandardType {
	case sc.G45_AT_TYPE, sc.G45_FAT_TYPE, sc.G45_NFT_TYPE:
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return p.g45DisplayContainer.Layout(gtx, th)
		})
	}

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return p.balanceContainer.Layout(gtx, th)
	})

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return p.sendReceiveButtons.Layout(gtx, th)
	})

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		text := make(map[string]string)
		text["txs"] = lang.Translate("Transactions")
		text["info"] = lang.Translate("Info")
		p.tabBars.Colors = theme.Current.TabBarsColors
		return p.tabBars.Layout(gtx, th, unit.Sp(18), text)
	})

	if p.tabBars.Key == "txs" {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return p.txBar.Layout(gtx, th)
		})

		if len(p.txItems) == 0 {
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th, unit.Sp(16), lang.Translate("You don't have any txs. Try adjusting filtering options or wait for wallet to sync."))
				lbl.Color = theme.Current.TextMuteColor
				return lbl.Layout(gtx)
			})
		}

		for i := range p.txItems {
			idx := i
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				return p.txItems[idx].Layout(gtx, th)
			})
		}
	}

	if p.tabBars.Key == "info" {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return p.tokenInfo.Layout(gtx, th)
		})
	}

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
	})

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(0), Bottom: unit.Dp(20),
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, widgets[index])
	})
}

type TokenInfoList struct {
	token    *wallet_manager.Token
	infoRows []*prefabs.InfoRow
}

func NewTokenInfoList(token *wallet_manager.Token) *TokenInfoList {
	return &TokenInfoList{
		token:    token,
		infoRows: prefabs.NewInfoRows(5),
	}
}

func (t *TokenInfoList) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	token := t.token
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.infoRows[0].Layout(gtx, th, lang.Translate("Name"), token.Name)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.infoRows[1].Layout(gtx, th, lang.Translate("Decimals"), fmt.Sprint(token.Decimals))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.infoRows[2].Layout(gtx, th, lang.Translate("Symbol"), token.Symbol.String)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			txt := "?"
			if token.MaxSupply.Valid {
				maxSupply := utils.ShiftNumber{Number: uint64(token.MaxSupply.Int64), Decimals: int(token.Decimals)}
				txt = maxSupply.Format()
			}

			return t.infoRows[3].Layout(gtx, th, lang.Translate("Max Supply"), txt)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.infoRows[4].Layout(gtx, th, lang.Translate("SC Standard"), fmt.Sprint(token.StandardType))
		}),
	)
}

type BalanceContainer struct {
	token             *wallet_manager.Token
	balanceEditor     *widget.Editor
	buttonHideBalance *ButtonHideBalance
	tokenImage        *prefabs.ImageHoverClick
}

func NewBalanceContainer() *BalanceContainer {
	buttonHideBalance := NewButtonHideBalance()
	balanceEditor := new(widget.Editor)
	balanceEditor.ReadOnly = true
	balanceEditor.SingleLine = true

	return &BalanceContainer{
		buttonHideBalance: buttonHideBalance,
		balanceEditor:     balanceEditor,
		tokenImage:        prefabs.NewImageHoverClick(),
	}
}

func (b *BalanceContainer) SetToken(token *wallet_manager.Token) {
	b.token = token
}

func (b *BalanceContainer) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if b.tokenImage.Clickable.Clicked() {
		image_modal.Instance.Open(b.token.Name, b.tokenImage.Image.Src)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			r := op.Record(gtx.Ops)
			dims := layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								txt := utils.ReduceTxId(b.token.SCID)
								if b.token.Symbol.String != "" {
									txt += fmt.Sprintf(" (%s)", b.token.Symbol.String)
								}

								lbl := material.Label(th, unit.Sp(14), txt)
								lbl.Color = theme.Current.TextMuteColor
								return lbl.Layout(gtx)
							}),
							//layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Axis:      layout.Horizontal,
									Alignment: layout.Middle,
								}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										b.tokenImage.Image.Src = b.token.LoadImageOp()
										gtx.Constraints.Max.X = gtx.Dp(35)
										gtx.Constraints.Max.Y = gtx.Dp(35)
										return b.tokenImage.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										wallet := wallet_manager.OpenedWallet
										balance, _ := wallet.Memory.Get_Balance_scid(b.token.GetHash())
										amount := utils.ShiftNumber{Number: balance, Decimals: int(b.token.Decimals)}.Format()

										if b.balanceEditor.Text() != amount {
											b.balanceEditor.SetText(amount)
										}

										r := op.Record(gtx.Ops)
										balanceEditor := material.Editor(th, b.balanceEditor, "")
										balanceEditor.TextSize = unit.Sp(17)
										balanceEditor.Font.Weight = font.Bold

										dims := balanceEditor.Layout(gtx)
										c := r.Stop()

										if settings.App.HideBalance {
											paint.FillShape(gtx.Ops, theme.Current.HideBalanceBgColor, clip.Rect{
												Max: dims.Size,
											}.Op())
										} else {
											c.Add(gtx.Ops)
										}

										return dims
									}),
								)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.Y = gtx.Dp(30)
						gtx.Constraints.Min.X = gtx.Dp(30)
						b.buttonHideBalance.Button.Style.Colors = theme.Current.ButtonIconPrimaryColors
						return b.buttonHideBalance.Layout(gtx, th)
					}),
				)
			})
			c := r.Stop()

			paint.FillShape(gtx.Ops, theme.Current.ListBgColor,
				clip.UniformRRect(
					image.Rectangle{Max: dims.Size},
					gtx.Dp(15),
				).Op(gtx.Ops))

			c.Add(gtx.Ops)
			return dims
		}),
	)
}

type G45DisplayContainer struct {
	token *wallet_manager.Token

	ownerEditor  *widget.Editor
	amountEditor *widget.Editor
}

func NewG45DisplayContainer() *G45DisplayContainer {
	ownerEditor := new(widget.Editor)
	ownerEditor.ReadOnly = true

	amountEditor := new(widget.Editor)
	amountEditor.ReadOnly = true

	return &G45DisplayContainer{
		ownerEditor:  ownerEditor,
		amountEditor: amountEditor,
	}
}

func (d *G45DisplayContainer) SetToken(token *wallet_manager.Token) {
	d.token = token
}

func (d *G45DisplayContainer) Load() {
	switch d.token.StandardType {
	case sc.G45_NFT_TYPE:
		d.ownerEditor.SetText("")

		var result rpc.GetSC_Result
		err := wallet_manager.RPC_Client.RPC.CallResult(context.Background(), "DERO.GetSC", rpc.GetSC_Params{
			SCID:       d.token.SCID,
			Code:       false,
			Variables:  false,
			KeysString: []string{"owner"},
		}, &result)
		if err != nil {
			d.ownerEditor.SetText("--")
			return
		}

		owner, err := utils.DecodeString(result.ValuesString[0])
		if err != nil {
			d.ownerEditor.SetText("--")
			return
		}

		d.ownerEditor.SetText(owner)
	case sc.G45_AT_TYPE, sc.G45_FAT_TYPE:
		d.amountEditor.SetText("")

		wallet := wallet_manager.OpenedWallet
		addr := wallet.Memory.GetAddress().String()
		key := fmt.Sprintf("owner_%s", addr)

		var result rpc.GetSC_Result
		err := wallet_manager.RPC_Client.RPC.CallResult(context.Background(), "DERO.GetSC", rpc.GetSC_Params{
			SCID:       d.token.SCID,
			Code:       false,
			Variables:  false,
			KeysString: []string{key},
		}, &result)
		if err != nil {
			d.amountEditor.SetText("--")
			return
		}

		amountDisplayed, err := strconv.ParseUint(result.ValuesString[0], 10, 64)
		if err != nil {
			d.amountEditor.SetText("--")
			return
		}

		amount := utils.ShiftNumber{Number: amountDisplayed, Decimals: int(d.token.Decimals)}
		d.amountEditor.SetText(amount.Format())
	}
}

func (d *G45DisplayContainer) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if d.token == nil { //|| d.amountDisplayed == 0 {
		return layout.Dimensions{}
	}

	r := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if d.token.StandardType == sc.G45_NFT_TYPE {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(14), lang.Translate("Owner"))
					lbl.Color = theme.Current.TextMuteColor
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if d.ownerEditor.Text() == "" {
						d.ownerEditor.SetText(lang.Translate("unknown"))
					}

					editor := material.Editor(th, d.ownerEditor, "")
					editor.Font.Weight = font.Bold
					editor.TextSize = unit.Sp(16)
					return editor.Layout(gtx)
				}),
			)
		}

		if d.token.StandardType == sc.G45_AT_TYPE ||
			d.token.StandardType == sc.G45_FAT_TYPE {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(14), lang.Translate("Amount Displayed"))
					lbl.Color = theme.Current.TextMuteColor
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					editor := material.Editor(th, d.amountEditor, "")
					editor.Font.Weight = font.Bold
					editor.TextSize = unit.Sp(20)
					return editor.Layout(gtx)
				}),
			)
		}

		return layout.Dimensions{}
	})
	c := r.Stop()

	paint.FillShape(gtx.Ops, theme.Current.ListBgColor,
		clip.UniformRRect(
			image.Rectangle{Max: dims.Size},
			gtx.Dp(15),
		).Op(gtx.Ops))

	c.Add(gtx.Ops)
	return dims
}
