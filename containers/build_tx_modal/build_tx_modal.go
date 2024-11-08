package build_tx_modal

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/transaction"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/app_instance"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/containers/password_modal"
	"github.com/secretsystems/secret-wallet/containers/recent_txs_modal"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/secretsystems/secret-wallet/utils"
	"github.com/secretsystems/secret-wallet/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type TxPayload struct {
	Transfers  []rpc.Transfer
	Ringsize   uint64
	SCArgs     rpc.Arguments
	TokensInfo []*wallet_manager.Token
}

func (t TxPayload) GetTokenInfo(scId crypto.Hash) *wallet_manager.Token {
	for _, asset := range t.TokensInfo {
		if crypto.HashHexToHash(asset.SCID) == scId {
			return asset
		}
	}

	return nil
}

func (t TxPayload) TotalDeroAmount() uint64 {
	totalDero := uint64(0)
	for _, transfer := range t.Transfers {
		if transfer.SCID.IsZero() {
			totalDero += transfer.Amount + transfer.Burn
		}
	}

	return totalDero
}

func (t TxPayload) TotalTokensAmount() map[crypto.Hash]uint64 {
	tokensAmount := make(map[crypto.Hash]uint64)

	for _, transfer := range t.Transfers {
		if !transfer.SCID.IsZero() {
			_, ok := tokensAmount[transfer.SCID]
			if !ok {
				tokensAmount[transfer.SCID] = 0
			}

			tokensAmount[transfer.SCID] += transfer.Amount + transfer.Burn
		}
	}

	return tokensAmount
}

type BuildTxModal struct {
	modal            *components.Modal
	buttonSend       *components.Button
	loadingIcon      *widget.Icon
	animationLoading *animation.Animation
	buttonClose      *components.Button

	building bool
	builtTx  *transaction.Transaction
	txFees   uint64
	gasFees  uint64
	txSent   bool

	txPayload TxPayload
}

var Instance *BuildTxModal

func LoadInstance() {
	modal := components.NewModal(components.ModalStyle{
		CloseOnOutsideClick: true,
		CloseOnInsideClick:  false,
		Direction:           layout.N,
		Inset:               layout.UniformInset(unit.Dp(10)),
		Rounded:             components.UniformRounded(unit.Dp(10)),
		Animation:           components.NewModalAnimationDown(),
	})

	sendIcon, _ := widget.NewIcon(icons.ContentSend)
	loadingIcon, _ := widget.NewIcon(icons.NavigationRefresh)
	buttonSend := components.NewButton(components.ButtonStyle{
		Rounded:     components.UniformRounded(unit.Dp(5)),
		Icon:        sendIcon,
		TextSize:    unit.Sp(14),
		IconGap:     unit.Dp(10),
		Inset:       layout.UniformInset(unit.Dp(10)),
		Animation:   components.NewButtonAnimationDefault(),
		LoadingIcon: loadingIcon,
	})
	buttonSend.Label.Alignment = text.Middle
	buttonSend.Style.Font.Weight = font.Bold

	closeIcon, _ := widget.NewIcon(icons.NavigationCancel)
	buttonClose := components.NewButton(components.ButtonStyle{
		Icon:      closeIcon,
		Animation: components.NewButtonAnimationDefault(),
	})

	editorError := new(widget.Editor)
	editorError.WrapPolicy = text.WrapGraphemes
	editorError.ReadOnly = true

	animationLoading := animation.NewAnimation(false,
		gween.NewSequence(
			gween.New(0, 1, 1, ease.Linear),
		),
	)
	animationLoading.Sequence.SetLoop(-1)

	Instance = &BuildTxModal{
		modal:            modal,
		buttonSend:       buttonSend,
		loadingIcon:      loadingIcon,
		animationLoading: animationLoading,
		buttonClose:      buttonClose,
	}

	app_instance.Router.AddLayout(router.KeyLayout{
		DrawIndex: 2,
		Layout:    Instance.layout,
	})
}

func (b *BuildTxModal) OpenWithRandomAddr(scId crypto.Hash, onLoad func(addr string, open func(txPayload TxPayload))) {
	wallet := wallet_manager.OpenedWallet
	b.modal.SetVisible(true)
	b.animationLoading.Reset().Start()
	b.building = true

	randomAddr, err := wallet.GetRandomAddress(scId)
	if err != nil {
		b.modal.SetVisible(false)
		notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
		notification_modals.ErrorInstance.SetVisible(true, 0)
	}

	onLoad(randomAddr, func(txPayload TxPayload) {
		b.Open(txPayload)
	})
}

func (b *BuildTxModal) Open(txPayload TxPayload) {
	wallet := wallet_manager.OpenedWallet
	b.txSent = false
	b.txPayload = txPayload
	b.builtTx = nil

	b.modal.SetVisible(true)
	b.animationLoading.Reset().Start()
	b.building = true

	tx, _, gasFees, err := wallet.BuildTransaction(txPayload.Transfers, txPayload.Ringsize, txPayload.SCArgs, false)
	b.animationLoading.Pause()

	if err != nil {
		b.modal.SetVisible(false)
		notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
		notification_modals.ErrorInstance.SetVisible(true, 0)
	} else {
		b.building = false
		b.builtTx = tx
		b.gasFees = gasFees
		b.txFees = tx.Fees()
	}
}

func (b *BuildTxModal) TxSent() bool {
	if b.txSent {
		b.txSent = false
		return true
	}

	return false
}

func (b *BuildTxModal) sendTx() error {
	b.buttonSend.SetLoading(true)
	wallet := wallet_manager.OpenedWallet

	err := wallet.InsertOutgoingTx(b.builtTx)
	if err != nil {
		b.buttonSend.SetLoading(false)
		return err
	}

	err = wallet.Memory.SendTransaction(b.builtTx)
	if err != nil {
		b.buttonSend.SetLoading(false)
		return err
	}

	b.buttonSend.SetLoading(false)
	b.modal.SetVisible(false)
	recent_txs_modal.Instance.SetVisible(true)
	b.txSent = true
	return nil
}

func (b *BuildTxModal) layout(gtx layout.Context, th *material.Theme) {
	wallet := wallet_manager.OpenedWallet

	if b.buttonSend.Clicked() {
		password_modal.Instance.SetVisible(true)
	}

	if b.buttonClose.Clicked() {
		b.modal.SetVisible(false)
	}

	submitted, password := password_modal.Instance.Input.Submitted()
	if submitted {
		validPassword := wallet.Memory.Check_Password(password)

		if !validPassword {
			password_modal.Instance.StartWrongPassAnimation()
		} else {
			err := b.sendTx()
			if err != nil {
				notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
				notification_modals.ErrorInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
			} else {
				password_modal.Instance.SetVisible(false)
			}
		}
	}

	b.modal.Style.Colors = theme.Current.ModalColors
	b.modal.Layout(gtx, nil, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(15), Bottom: unit.Dp(15),
			Left: unit.Dp(15), Right: unit.Dp(15),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			var childs []layout.FlexChild

			if b.building {
				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(20), lang.Translate("Sending transaction..."))
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								r := op.Record(gtx.Ops)
								dims := b.loadingIcon.Layout(gtx, th.Fg)
								c := r.Stop()

								{
									gtx.Constraints.Min = dims.Size

									state := b.animationLoading.Update(gtx)
									if state.Active {
										defer animation.TransformRotate(gtx, state.Value).Push(gtx.Ops).Pop()
									}
								}

								c.Add(gtx.Ops)
								return dims
							}),
						)
					}))
			} else if b.builtTx != nil {
				totalDero := b.txPayload.TotalDeroAmount()

				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(22), lang.Translate("Confirm"))
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								b.buttonClose.Style.Colors = theme.Current.ModalButtonColors
								return b.buttonClose.Layout(gtx, th)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				)

				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(16), lang.Translate("Ring size"))
								lbl.Color = theme.Current.TextMuteColor
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(16), fmt.Sprint(b.txPayload.Ringsize))
								return lbl.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
				)

				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(16), lang.Translate("Transfer"))
								lbl.Color = theme.Current.TextMuteColor
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("%s DERO", globals.FormatMoney(totalDero)))
								return lbl.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						var flexChilds []layout.FlexChild

						tokensAmount := b.txPayload.TotalTokensAmount()
						for scId, amount := range tokensAmount {
							amountString := fmt.Sprint(amount)
							assetString := utils.ReduceTxId(scId.String())

							token := b.txPayload.GetTokenInfo(scId)
							if token != nil {
								if token.Name != "" {
									assetString += fmt.Sprintf(" (%s)", token.Name)
								}

								amountString = utils.ShiftNumber{Number: amount, Decimals: int(token.Decimals)}.Format()
								if token.Symbol.Valid {
									amountString += fmt.Sprintf(" %s", token.Symbol.String)
								}
							}

							flexChilds = append(flexChilds, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(th, unit.Sp(14), assetString)
										lbl.Color = theme.Current.TextMuteColor
										lbl.Alignment = text.End
										return lbl.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Width: unit.Dp(5)}.Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(th, unit.Sp(14), amountString)
										return lbl.Layout(gtx)
									}),
								)
							}))
						}

						return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChilds...)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
				)

				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(16), lang.Translate("TX fees"))
								lbl.Color = theme.Current.TextMuteColor
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								fees := globals.FormatMoney(b.txFees)
								lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("%s DERO", fees))
								return lbl.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
				)

				if len(b.txPayload.Transfers) >= 1 && len(b.txPayload.SCArgs) == 0 {
					childs = append(childs,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(th, unit.Sp(16), lang.Translate("Receiver"))
									lbl.Color = theme.Current.TextMuteColor
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									txt := ""
									if len(b.txPayload.Transfers) > 1 {
										txt = lang.Translate("Multiple receivers")
									} else if len(b.txPayload.Transfers) == 1 {
										addr := b.txPayload.Transfers[0].Destination
										txt = utils.ReduceAddr(addr)
									}

									lbl := material.Label(th, unit.Sp(16), txt)
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
					)
				}

				if b.txPayload.SCArgs.HasValue(rpc.SCID, rpc.DataHash) {
					childs = append(childs,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(th, unit.Sp(16), lang.Translate("SC Call"))
									lbl.Color = theme.Current.TextMuteColor
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									scId := b.txPayload.SCArgs.Value(rpc.SCID, rpc.DataHash).(crypto.Hash)
									txt := utils.ReduceAddr(scId.String())
									lbl := material.Label(th, unit.Sp(16), txt)
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
					)
				}

				if b.txPayload.SCArgs.HasValue("entrypoint", rpc.DataString) {
					childs = append(childs,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(th, unit.Sp(16), lang.Translate("Entrypoint"))
									lbl.Color = theme.Current.TextMuteColor
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									entrypoint := b.txPayload.SCArgs.Value("entrypoint", rpc.DataString).(string)
									lbl := material.Label(th, unit.Sp(16), entrypoint)
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
					)

					childs = append(childs,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(th, unit.Sp(16), lang.Translate("Gas fees"))
									lbl.Color = theme.Current.TextMuteColor
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									fees := globals.FormatMoney(b.gasFees)
									lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("%s DERO", fees))
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
					)
				}

				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(th, unit.Sp(16), lang.Translate("Total"))
								lbl.Color = theme.Current.TextMuteColor
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								total := globals.FormatMoney(totalDero + b.txFees + b.gasFees)
								lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("%s DERO", total))
								return lbl.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						b.buttonSend.Text = lang.Translate("SEND TRANSACTION")
						b.buttonSend.Style.Colors = theme.Current.ButtonPrimaryColors
						return b.buttonSend.Layout(gtx, th)
					}),
				)
			}

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, childs...)
		})
	})
}
