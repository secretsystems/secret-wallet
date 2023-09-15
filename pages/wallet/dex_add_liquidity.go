package page_wallet

import (
	"fmt"
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	crypto "github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/g45w/animation"
	"github.com/g45t345rt/g45w/components"
	"github.com/g45t345rt/g45w/containers/build_tx_modal"
	"github.com/g45t345rt/g45w/containers/notification_modals"
	"github.com/g45t345rt/g45w/lang"
	"github.com/g45t345rt/g45w/router"
	"github.com/g45t345rt/g45w/sc/dex_sc"
	"github.com/g45t345rt/g45w/theme"
	"github.com/g45t345rt/g45w/utils"
	"github.com/g45t345rt/g45w/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageDEXAddLiquidity struct {
	isActive bool

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	buttonAdd               *components.Button
	liquidityContainer      *LiquidityContainer
	pairTokenInputContainer *PairTokenInputContainer

	pair dex_sc.Pair

	list *widget.List
}

var _ router.Page = &PageDEXAddLiquidity{}

func NewPageDEXAddLiquidity() *PageDEXAddLiquidity {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, 1, .25, ease.Linear),
	))

	list := new(widget.List)
	list.Axis = layout.Vertical

	addIcon, _ := widget.NewIcon(icons.ContentAdd)
	buttonAdd := components.NewButton(components.ButtonStyle{
		Rounded:   components.UniformRounded(unit.Dp(5)),
		Icon:      addIcon,
		TextSize:  unit.Sp(14),
		IconGap:   unit.Dp(10),
		Inset:     layout.UniformInset(unit.Dp(10)),
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonAdd.Label.Alignment = text.Middle
	buttonAdd.Style.Font.Weight = font.Bold

	pairTokenInputContainer := NewPairTokenInputContainer()

	return &PageDEXAddLiquidity{
		animationEnter:          animationEnter,
		animationLeave:          animationLeave,
		list:                    list,
		pairTokenInputContainer: pairTokenInputContainer,
		buttonAdd:               buttonAdd,
		liquidityContainer:      NewLiquidityContainer(),
	}
}

func (p *PageDEXAddLiquidity) IsActive() bool {
	return p.isActive
}

func (p *PageDEXAddLiquidity) Enter() {
	p.isActive = true

	if !page_instance.header.IsHistory(PAGE_DEX_ADD_LIQUIDITY) {
		p.animationEnter.Start()
		p.animationLeave.Reset()
	}
	page_instance.header.Title = func() string {
		return lang.Translate("Add Liquidity")
	}

	page_instance.header.Subtitle = func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		lbl := material.Label(th, unit.Sp(14), p.pair.Symbol)
		lbl.Color = theme.Current.TextMuteColor
		return lbl.Layout(gtx)
	}

	page_instance.header.ButtonRight = nil
}

func (p *PageDEXAddLiquidity) SetPair(pair dex_sc.Pair, token1 *wallet_manager.Token, token2 *wallet_manager.Token) {
	p.pair = pair
	p.liquidityContainer.SetPair(pair, token1, token2)
	p.pairTokenInputContainer.SetTokens(token1, token2)

	if pair.SharesOutstanding > 0 {
		p.pairTokenInputContainer.txtAmount2.Editor().ReadOnly = true
	} else {
		p.pairTokenInputContainer.txtAmount2.Editor().ReadOnly = false
	}
}

func (p *PageDEXAddLiquidity) Leave() {
	p.animationLeave.Start()
	p.animationEnter.Reset()
}

func (p *PageDEXAddLiquidity) submitForm() error {
	token1 := p.pairTokenInputContainer.token1
	token2 := p.pairTokenInputContainer.token2

	build_tx_modal.Instance.OpenWithRandomAddr(crypto.ZEROHASH, func(addr string, open func(txPayload build_tx_modal.TxPayload)) {
		open(build_tx_modal.TxPayload{
			SCArgs: rpc.Arguments{
				{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
				{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(p.pair.SCID)},
				{Name: "entrypoint", DataType: rpc.DataString, Value: "AddLiquidity"},
			},
			Transfers: []rpc.Transfer{
				rpc.Transfer{SCID: token1.GetHash(), Burn: 0, Destination: addr},
				rpc.Transfer{SCID: token2.GetHash(), Burn: 0, Destination: addr},
			},
			Ringsize:   2,
			TokensInfo: []*wallet_manager.Token{token1, token2},
		})
	})

	return nil
}

func (p *PageDEXAddLiquidity) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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

	if p.buttonAdd.Clicked() {
		go func() {
			err := p.submitForm()
			if err != nil {
				notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
				notification_modals.ErrorInstance.SetVisible(true, 0)
			}
		}()
	}

	widgets := []layout.Widget{}

	if p.pair.SharesOutstanding == 0 {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(16), lang.Translate("Looks like there is no shares. You will provide the initial liquidity to the pair."))
			lbl.Color = theme.Current.TextMuteColor
			return lbl.Layout(gtx)
		})
	} else {
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return p.liquidityContainer.Layout(gtx, th)
		})
	}

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return p.pairTokenInputContainer.Layout(gtx, th, lang.Translate("SEND {}"), lang.Translate("SEND {}"))
	})

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		p.buttonAdd.Style.Colors = theme.Current.ButtonPrimaryColors
		p.buttonAdd.Text = lang.Translate("ADD")
		return p.buttonAdd.Layout(gtx, th)
	})

	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
	})

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Overlay

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Inset{
			Bottom: unit.Dp(20),
			Left:   unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return widgets[index](gtx)
		})
	})
}

type LiquidityContainer struct {
	pair   dex_sc.Pair
	token1 *wallet_manager.Token
	token2 *wallet_manager.Token
	share  uint64
}

func NewLiquidityContainer() *LiquidityContainer {
	return &LiquidityContainer{}
}

func (p *LiquidityContainer) SetPair(pair dex_sc.Pair, token1 *wallet_manager.Token, token2 *wallet_manager.Token) {
	p.pair = pair
	p.token1 = token1
	p.token2 = token2

	go func() {
		wallet := wallet_manager.OpenedWallet
		addr := wallet.Memory.GetAddress().String()
		p.share, _, _ = wallet.Memory.GetDecryptedBalanceAtTopoHeight(crypto.HashHexToHash(p.pair.SCID), -1, addr)
	}()
}

func (p *LiquidityContainer) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	r := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th, unit.Sp(18), lang.Translate("Your liquidity"))
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						share := p.pair.CalcShare(p.pair.Asset1, p.share)
						amount := utils.ShiftNumber{Number: share, Decimals: int(p.token1.Decimals)}
						lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("%s %s", amount.Format(), p.token1.Symbol.String))
						lbl.Color = theme.Current.TextMuteColor
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(1)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						share := p.pair.CalcShare(p.pair.Asset2, p.share)
						amount := utils.ShiftNumber{Number: share, Decimals: int(p.token2.Decimals)}
						lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("%s %s", amount.Format(), p.token2.Symbol.String))
						lbl.Color = theme.Current.TextMuteColor
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(1)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						value := p.pair.CalcOwnership(p.share)
						lbl := material.Label(th, unit.Sp(16), fmt.Sprintf("Ownership: %.3f", value))
						lbl.Color = theme.Current.TextMuteColor
						return lbl.Layout(gtx)
					}),
				)
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
}