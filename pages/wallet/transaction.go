package page_wallet

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	"gioui.org/font"
	"gioui.org/io/clipboard"
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
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/prefabs"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/secretsystems/secret-wallet/utils"
	"github.com/secretsystems/secret-wallet/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageTransaction struct {
	isActive bool
	list     *widget.List

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	infoRows []*prefabs.InfoRow

	entry wallet_manager.Entry

	txTypeImg components.Image

	txIdEditor              *widget.Editor
	senderDestinationEditor *widget.Editor
	blockHashEditor         *widget.Editor
	proofEditor             *widget.Editor
	scDataEditor            *widget.Editor

	txTransfers *TxTransfers

	payloadList []*RPCArgInfo
}

var _ router.Page = &PageTransaction{}

func NewPageTransaction() *PageTransaction {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .25, ease.Linear),
	))

	txTypeImg := components.Image{
		Fit: components.Cover,
	}

	list := new(widget.List)
	list.Axis = layout.Vertical

	return &PageTransaction{
		animationEnter: animationEnter,
		animationLeave: animationLeave,

		list:                    list,
		txIdEditor:              &widget.Editor{ReadOnly: true},
		senderDestinationEditor: &widget.Editor{ReadOnly: true},
		blockHashEditor:         &widget.Editor{ReadOnly: true},
		proofEditor:             &widget.Editor{ReadOnly: true},
		scDataEditor:            &widget.Editor{ReadOnly: true},
		infoRows:                prefabs.NewInfoRows(3),
		txTransfers:             NewTxTransfers(),

		txTypeImg: txTypeImg,
	}
}

func (p *PageTransaction) IsActive() bool {
	return p.isActive
}

func (p *PageTransaction) Clear() {
	p.payloadList = make([]*RPCArgInfo, 0)
	p.txIdEditor.SetText("")
	p.senderDestinationEditor.SetText("")
	p.blockHashEditor.SetText("")
	p.proofEditor.SetText("")
}

func (p *PageTransaction) Enter() {
	p.Clear()
	p.txIdEditor.SetText(p.entry.TXID)

	for _, arg := range p.entry.Payload_RPC {
		p.payloadList = append(p.payloadList, NewRPCArgInfo(arg))
	}

	scData, _ := json.MarshalIndent(p.entry.Payload_RPC, "", "  ")
	p.scDataEditor.SetText(string(scData))

	if p.entry.Incoming {
		sender := p.entry.Sender
		if sender == "" {
			sender = "?"
		}

		p.senderDestinationEditor.SetText(sender)
		p.txTypeImg.Src = theme.Current.ArrowDownArcImage
	} else {
		p.senderDestinationEditor.SetText(p.entry.Destination)
		p.txTypeImg.Src = theme.Current.ArrowUpArcImage
	}

	p.blockHashEditor.SetText(p.entry.BlockHash)
	p.proofEditor.SetText(p.entry.Proof)

	page_instance.header.Title = func() string { return lang.Translate("Transaction") }
	page_instance.header.Subtitle = func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		txId := utils.ReduceTxId(p.entry.TXID)
		if txId == "" {
			txId = lang.Translate("From Coinbase")
		}

		lbl := material.Label(th, unit.Sp(16), txId)
		lbl.Color = theme.Current.TextMuteColor
		return lbl.Layout(gtx)
	}
	page_instance.header.ButtonRight = nil

	p.isActive = true
	if !page_instance.header.IsHistory(PAGE_TRANSACTION) {
		p.animationEnter.Start()
		p.animationLeave.Reset()
	}
}

func (p *PageTransaction) Leave() {
	p.animationLeave.Start()
	p.animationEnter.Reset()
}

func (p *PageTransaction) SetEntry(e wallet_manager.Entry) {
	p.entry = e
	p.txTransfers.Load(e)
}

func (p *PageTransaction) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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

	widgets := []layout.Widget{}

	if p.entry.TXID != "" {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					r := op.Record(gtx.Ops)
					dims := layout.UniformInset(unit.Dp(15)).Layout(
						gtx,
						func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis:      layout.Horizontal,
								Alignment: layout.Middle,
							}.Layout(
								gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Max.X = gtx.Dp(50)
									gtx.Constraints.Max.Y = gtx.Dp(50)
									return p.txTypeImg.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									editor := material.Editor(th, p.txIdEditor, "")
									return editor.Layout(gtx)
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
		})
	}

	if !p.entry.Coinbase {
		if p.entry.Incoming {
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th, unit.Sp(16), lang.Translate("Sender"))
						lbl.Font.Weight = font.Bold
						lbl.Color = theme.Current.TextMuteColor
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						editor := material.Editor(th, p.senderDestinationEditor, "")
						return editor.Layout(gtx)
					}),
				)
			})
		} else {
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th, unit.Sp(16), lang.Translate("Destination"))
						lbl.Font.Weight = font.Bold
						lbl.Color = theme.Current.TextMuteColor
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						editor := material.Editor(th, p.senderDestinationEditor, "")
						return editor.Layout(gtx)
					}),
				)
			})
		}
	}

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				blockHeight := fmt.Sprint(p.entry.Height)
				return p.infoRows[0].Layout(gtx, th, lang.Translate("Block Height"), blockHeight)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				date := p.entry.Time.Format("2006-01-02 15:04")
				return p.infoRows[1].Layout(gtx, th, lang.Translate("Date"), date)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				timeAgo := lang.TimeAgo(p.entry.Time)
				return p.infoRows[2].Layout(gtx, th, lang.Translate("Time"), timeAgo)
			}),
		)
	})

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return p.txTransfers.Layout(gtx, th)
	})

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th, unit.Sp(16), lang.Translate("Block Hash"))
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.Current.TextMuteColor
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				editor := material.Editor(th, p.blockHashEditor, "")
				return editor.Layout(gtx)
			}),
		)
	})
	// to do
	// we need to break out the tx proof structure
	// basically what was done for the tx comment
	// good luck
	if !p.entry.Coinbase {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(16), lang.Translate("Proof"))
					lbl.Font.Weight = font.Bold
					lbl.Color = theme.Current.TextMuteColor
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					editor := material.Editor(th, p.proofEditor, "")
					return editor.Layout(gtx)
				}),
			)
		})
	}

	for i := range p.payloadList {
		idx := i
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return p.payloadList[idx].Layout(gtx, th)
		})
	}

	if len(p.entry.Payload_RPC) > 0 {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			var childs []layout.FlexChild

			childs = append(childs,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(16), "SC DATA")
					lbl.Color = theme.Current.TextMuteColor
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					r := op.Record(gtx.Ops)
					dims := layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						editor := material.Editor(th, p.scDataEditor, "")
						editor.TextSize = unit.Sp(12)
						return editor.Layout(gtx)
					})
					c := r.Stop()

					paint.FillShape(gtx.Ops, theme.Current.BgColor, clip.RRect{
						Rect: image.Rectangle{Max: dims.Size},
						NW:   gtx.Dp(10), NE: gtx.Dp(10),
						SE: gtx.Dp(10), SW: gtx.Dp(10),
					}.Op(gtx.Ops))

					c.Add(gtx.Ops)
					return dims
				}),
			)

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, childs...)
		})
	}

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
	})

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Occupy

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(0), Bottom: unit.Dp(20),
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, widgets[index])
	})
}

type RPCArgInfo struct {
	arg        rpc.Argument
	editor     *widget.Editor
	buttonCopy *components.Button
}

func NewRPCArgInfo(arg rpc.Argument) *RPCArgInfo {
	editor := new(widget.Editor)
	editor.ReadOnly = true
	editor.SetText(fmt.Sprint(arg.Value))

	copyIcon, _ := widget.NewIcon(icons.ContentContentCopy)
	buttonCopy := components.NewButton(components.ButtonStyle{
		Icon: copyIcon,
	})

	return &RPCArgInfo{
		editor:     editor,
		arg:        arg,
		buttonCopy: buttonCopy,
	}
}

func (p *RPCArgInfo) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.buttonCopy.Clicked() {
		clipboard.WriteOp{
			Text: p.editor.Text(),
		}.Add(gtx.Ops)
		notification_modals.InfoInstance.SetText(lang.Translate("Clipboard"), lang.Translate("Text copied to clipboard"))
		notification_modals.InfoInstance.SetVisible(true, notification_modals.CLOSE_AFTER_DEFAULT)
	}
	var name string
	switch p.arg.Name {
	case "D":
		name = lang.Translate("Destination Port")
	case "S":
		name = lang.Translate("Source Port")
	case "V":
		name = lang.Translate("Value Transfer")
	case "C":
		name = lang.Translate("Comment")
	case "E":
		name = lang.Translate("Expiry")
	case "R":
		name = lang.Translate("Replyback Address")
	case "N":
		name = lang.Translate("Needs Replyback Address")
	default:
		name = p.arg.Name
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(18), lang.Translate(name))
					lbl.Font.Weight = font.Bold
					lbl.Color = theme.Current.TextMuteColor
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = gtx.Dp(20)
					gtx.Constraints.Max.Y = gtx.Dp(20)
					p.buttonCopy.Style.Colors = theme.Current.ModalButtonColors
					return p.buttonCopy.Layout(gtx, th)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			editor := material.Editor(th, p.editor, "")
			editor.Editor.WrapPolicy = text.WrapHeuristically
			editor.Editor.SingleLine = 0 != 0
			return editor.Layout(gtx)
		}),
	)
}

type TxTransferItem struct {
	entry    wallet_manager.Entry
	token    *wallet_manager.Token
	infoRows []*prefabs.InfoRow
}

func NewTxTransferItem(entry wallet_manager.Entry) *TxTransferItem {
	token, _ := wallet_manager.GetTokenBySCID(entry.SCID.String())
	if token == nil {
		token = &wallet_manager.Token{
			Decimals: 0,
			SCID:     entry.SCID.String(),
		}
	}

	return &TxTransferItem{
		token:    token,
		entry:    entry,
		infoRows: prefabs.NewInfoRows(5),
	}
}

func (t *TxTransferItem) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			txt := utils.ReduceTxId(t.token.SCID)
			symbol := t.token.Symbol.String
			if symbol != "" {
				txt += fmt.Sprintf(" (%s)", symbol)
			}

			return t.infoRows[0].Layout(gtx, th, lang.Translate("SCID"), txt)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			amount := utils.ShiftNumber{Number: t.entry.Amount, Decimals: int(t.token.Decimals)}
			return t.infoRows[1].Layout(gtx, th, lang.Translate("Amount"), amount.Format())
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			fees := utils.ShiftNumber{Number: t.entry.Fees, Decimals: int(t.token.Decimals)}
			return t.infoRows[2].Layout(gtx, th, lang.Translate("Fees"), fees.Format())
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			burn := utils.ShiftNumber{Number: t.entry.Amount, Decimals: int(t.token.Decimals)}
			return t.infoRows[3].Layout(gtx, th, lang.Translate("Burn"), burn.Format())
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			txt := utils.ReduceAddr(t.entry.Destination)
			if txt == "" {
				txt = "?"
			}
			return t.infoRows[4].Layout(gtx, th, lang.Translate("Destination"), txt)
		}),
	)
}

type TxTransfers struct {
	items []*TxTransferItem
}

func NewTxTransfers() *TxTransfers {
	return &TxTransfers{
		items: make([]*TxTransferItem, 0),
	}
}

func (t *TxTransfers) Load(entry wallet_manager.Entry) {
	wallet := wallet_manager.OpenedWallet

	t.items = make([]*TxTransferItem, 0)
	if entry.TXID != "" {
		entries := wallet.GetEntries(nil, wallet_manager.GetEntriesParams{
			TXID: sql.NullString{String: entry.TXID, Valid: true},
		})

		for _, entry := range entries {
			t.items = append(t.items, NewTxTransferItem(entry))
		}
	} else {
		// from coinbase
		t.items = append(t.items, NewTxTransferItem(entry))
	}
}

func (t *TxTransfers) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			txt := lang.Translate("Transfers ({})")
			txt = strings.Replace(txt, "{}", fmt.Sprint(len(t.items)), -1)
			lbl := material.Label(th, unit.Sp(16), txt)
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var childs []layout.FlexChild
			for i := range t.items {
				idx := i
				childs = append(childs,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return t.items[idx].Layout(gtx, th)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				)
			}

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, childs...)
		}),
	)
}
