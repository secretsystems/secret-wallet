package page_wallet_select

import (
	"fmt"
	"image"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/app_db"
	"github.com/secretsystems/secret-wallet/app_instance"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/listselect_modal"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/containers/password_modal"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/pages"
	page_wallet "github.com/secretsystems/secret-wallet/pages/wallet"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/secretsystems/secret-wallet/utils"
	"github.com/secretsystems/secret-wallet/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageSelectWallet struct {
	isActive  bool
	clickable *widget.Clickable

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	buttonWalletCreate *components.Button
	walletList         *widget.List
	dragItems          *components.DragItems
	items              []walletItem

	currentWallet app_db.WalletInfo
}

var _ router.Page = &PageSelectWallet{}

func NewPageSelectWallet() *PageSelectWallet {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(-1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, -1, .25, ease.Linear),
	))

	//walletList := NewWalletList()

	addIcon, _ := widget.NewIcon(icons.ContentAddCircleOutline)
	buttonWalletCreate := components.NewButton(components.ButtonStyle{
		Rounded:   components.UniformRounded(unit.Dp(5)),
		Icon:      addIcon,
		TextSize:  unit.Sp(14),
		IconGap:   unit.Dp(10),
		Inset:     layout.UniformInset(unit.Dp(10)),
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonWalletCreate.Label.Alignment = text.Middle
	buttonWalletCreate.Style.Font.Weight = font.Bold

	walletList := new(widget.List)
	walletList.Axis = layout.Vertical
	dragItems := components.NewDragItems()

	return &PageSelectWallet{
		clickable: new(widget.Clickable),

		animationEnter: animationEnter,
		animationLeave: animationLeave,

		buttonWalletCreate: buttonWalletCreate,
		walletList:         walletList,
		dragItems:          dragItems,
	}
}

func (p *PageSelectWallet) IsActive() bool {
	return p.isActive
}

func (p *PageSelectWallet) Enter() {
	p.isActive = true
	page_instance.header.Title = func() string { return lang.Translate("Select wallet") }

	if !page_instance.header.IsHistory(PAGE_SELECT_WALLET) {
		p.animationLeave.Reset()
		p.animationEnter.Start()
	}

	p.Load()
}

func (p *PageSelectWallet) Leave() {
	p.animationEnter.Reset()
	p.animationLeave.Start()
}

func (p *PageSelectWallet) Load() error {
	p.items = make([]walletItem, 0)
	wallets, err := app_db.GetWallets()
	if err != nil {
		return err
	}

	for _, walletInfo := range wallets {
		p.items = append(p.items, walletItem{
			clickable:  new(widget.Clickable),
			walletInfo: walletInfo,
		})
	}

	return nil
}

func (p *PageSelectWallet) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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

	{
		moved, cIndex, nIndex := p.dragItems.ItemMoved()
		if moved {
			go func() {
				updateIndex := func() error {
					walletInfo := p.items[cIndex].walletInfo
					walletInfo.OrderNumber = nIndex
					err := app_db.UpdateWalletInfo(walletInfo)
					if err != nil {
						return err
					}

					return p.Load()
				}

				err := updateIndex()
				if err != nil {
					notification_modals.ErrorInstance.SetText("Error", err.Error())
					notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
				}
				app_instance.Window.Invalidate()
			}()
		}
	}

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(0), Bottom: unit.Dp(30),
				Left: unit.Dp(30), Right: unit.Dp(30),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if len(p.items) == 0 {
							labelNoWallet := material.Label(th, unit.Sp(16), lang.Translate("You didn't add a wallet yet.\nClick 'New Wallet' button to continue."))
							return labelNoWallet.Layout(gtx)
						} else {
							paint.FillShape(gtx.Ops, theme.Current.ListBgColor,
								clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Max}, gtx.Dp(10)).Op(gtx.Ops),
							)

							listStyle := material.List(th, p.walletList)
							listStyle.AnchorStrategy = material.Overlay
							listStyle.Indicator.MinorWidth = unit.Dp(10)
							listStyle.Indicator.CornerRadius = unit.Dp(5)
							listStyle.Indicator.Color = theme.Current.ListScrollBarBgColor

							return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.dragItems.Layout(gtx, &p.walletList.Position, func(gtx layout.Context) layout.Dimensions {
									return listStyle.Layout(gtx, len(p.items), func(gtx layout.Context, index int) layout.Dimensions {
										item := p.items[index]

										if item.clickable.Clicked() {
											p.currentWallet = item.walletInfo
											password_modal.Instance.SetVisible(true)
										}

										r := op.Record(gtx.Ops)
										dims := item.Layout(gtx, th, false)
										c := r.Stop()

										p.dragItems.LayoutItem(gtx, index, func(gtx layout.Context) layout.Dimensions {
											defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, 12).Push(gtx.Ops).Pop()
											return item.Layout(gtx, th, true)
										})

										return item.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{Bottom: unit.Dp(5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												c.Add(gtx.Ops)
												return dims
											})
										})
									})
								})
							})
						}
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(30)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if p.buttonWalletCreate.Clicked() {
							go func() {
								// fastIcon, _ := widget.NewIcon(icons.ImageFlashOn)
								newIcon, _ := widget.NewIcon(icons.ContentAddCircle)
								diskIcon, _ := widget.NewIcon(icons.FileFolder)
								seedIcon, _ := widget.NewIcon(icons.EditorShortText)

								keyChan := listselect_modal.Instance.Open([]*listselect_modal.SelectListItem{
									// listselect_modal.NewSelectListItem(PAGE_CREATE_WALLET_FASTREG_FORM,
									// 	listselect_modal.NewItemText(fastIcon, lang.Translate("Fast registration")).Layout,
									// ),
									listselect_modal.NewSelectListItem(PAGE_CREATE_WALLET_FORM,
										listselect_modal.NewItemText(newIcon, lang.Translate("Create new wallet")).Layout,
									),
									listselect_modal.NewSelectListItem(PAGE_CREATE_WALLET_DISK_FORM,
										listselect_modal.NewItemText(diskIcon, lang.Translate("Recover from disk")).Layout,
									),
									listselect_modal.NewSelectListItem(PAGE_CREATE_WALLET_SEED_FORM,
										listselect_modal.NewItemText(seedIcon, lang.Translate("Recover from seed")).Layout,
									),
									listselect_modal.NewSelectListItem(PAGE_CREATE_WALLET_HEXSEED_FORM,
										listselect_modal.NewItemText(seedIcon, lang.Translate("Recover from hex seed")).Layout,
									),
								})

								for key := range keyChan {
									page_instance.pageRouter.SetCurrent(key)
									page_instance.header.AddHistory(key)
								}
							}()
						}

						p.buttonWalletCreate.Text = lang.Translate("NEW WALLET")
						p.buttonWalletCreate.Style.Colors = theme.Current.ButtonPrimaryColors
						return p.buttonWalletCreate.Layout(gtx, th)
					}),
				)
			})
		}),
	)

	{
		submitted, text := password_modal.Instance.Input.Submitted()
		if submitted {
			go func() {
				password_modal.Instance.SetLoading(true)
				err := wallet_manager.OpenWallet(p.currentWallet.Addr, text)
				password_modal.Instance.SetLoading(false)
				if err == nil {
					wallet := wallet_manager.OpenedWallet
					wallet.Memory.SetOnlineMode()
					password_modal.Instance.SetVisible(false)
					// important reset wallet pages to initial state
					app_instance.Router.Pages[pages.PAGE_WALLET] = page_wallet.New()
					app_instance.Router.SetCurrent(pages.PAGE_WALLET)
				} else {
					if err.Error() == "Invalid Password" {
						password_modal.Instance.StartWrongPassAnimation()
					} else {
						//p.modalWalletPassword.Modal.SetVisible(false)
						notification_modals.ErrorInstance.SetText("Error", err.Error())
						notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
					}
				}
			}()
		}
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

type walletItem struct {
	clickable  *widget.Clickable
	walletInfo app_db.WalletInfo
}

func (item *walletItem) Layout(gtx layout.Context, th *material.Theme, fill bool) layout.Dimensions {
	r := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Start}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						name := fmt.Sprintf("%s [%s]", lang.Translate("Wallet"), item.walletInfo.Name)
						lbl := material.Label(th, unit.Sp(18), name)
						lbl.Font.Weight = font.Bold
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						addr := utils.ReduceAddr(item.walletInfo.Addr)
						lbl := material.Label(th, unit.Sp(15), addr)
						lbl.Color = theme.Current.TextMuteColor
						return lbl.Layout(gtx)
					}),
				)
			}),
		)
	})
	c := r.Stop()

	if item.clickable.Hovered() || fill {
		pointer.CursorPointer.Add(gtx.Ops)
		paint.FillShape(gtx.Ops, theme.Current.ListItemHoverBgColor,
			clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(12)).Op(gtx.Ops),
		)
	}

	c.Add(gtx.Ops)
	return dims
}
