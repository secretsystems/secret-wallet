package node_status_bar

import (
	"fmt"
	"image/color"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/app_instance"

	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/node_manager"
	"github.com/secretsystems/secret-wallet/pages"
	page_node "github.com/secretsystems/secret-wallet/pages/node"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/secretsystems/secret-wallet/wallet_manager"
)

type NodeStatusBar struct {
	clickable      *widget.Clickable
	RemoteNodeInfo *page_node.RemoteNodeInfo
}

var Instance *NodeStatusBar

func LoadInstance() *NodeStatusBar {
	nodeStatusBar := &NodeStatusBar{
		clickable:      new(widget.Clickable),
		RemoteNodeInfo: page_node.NewRemoteNodeInfo(3 * time.Second),
	}
	Instance = nodeStatusBar
	return nodeStatusBar
}

func (n *NodeStatusBar) Update() {
	currentNode := node_manager.CurrentNode
	if currentNode != nil {

		n.RemoteNodeInfo.Update()
	}
}

func (n *NodeStatusBar) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	bgColor := theme.Current.NodeStatusBgColor
	paint.FillShape(gtx.Ops, bgColor, clip.Rect{
		Max: gtx.Constraints.Max,
	}.Op())

	if wallet_manager.OpenedWallet == nil {
		return layout.Dimensions{}
	}

	wallet := wallet_manager.OpenedWallet
	currentNode := node_manager.CurrentNode
	status := "unassigned node"
	statusDotColor := theme.Current.NodeStatusDotRedColor // color.NRGBA{R: 255, G: 0, B: 0, A: 255}

	if currentNode != nil {

		n.RemoteNodeInfo.Active()
		walletHeight := wallet.Memory.Get_Height()
		daemonHeight := wallet.Memory.Get_Daemon_Height()
		out := n.RemoteNodeInfo.Result.Outgoing_connections_count

		if n.RemoteNodeInfo.Err == nil {
			if walletHeight < daemonHeight {
				statusDotColor = theme.Current.NodeStatusDotYellowColor //color.NRGBA{R: 255, G: 255, B: 0, A: 255}
			} else {
				statusDotColor = theme.Current.NodeStatusDotGreenColor // color.NRGBA{R: 0, G: 255, B: 0, A: 255}
			}

			status = fmt.Sprintf("%d / %d - %dP (%s)", walletHeight, daemonHeight, out, currentNode.Name)
		} else {
			status = fmt.Sprintf("%s (%s)", lang.Translate("Disconnected"), currentNode.Name)
		}

	}

	if n.clickable.Hovered() {
		pointer.CursorPointer.Add(gtx.Ops)
	}

	if n.clickable.Clicked() {
		app_instance.Router.SetCurrent(pages.PAGE_NODE)
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	return n.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(15), Bottom: unit.Dp(15),
			Left: unit.Dp(10), Right: unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return StatusDot{
						Color: statusDotColor,
					}.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(16), status)
					lbl.Color = theme.Current.NodeStatusTextColor
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

type StatusDot struct {
	Color color.NRGBA
}

func (s StatusDot) Layout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Max.X = gtx.Dp(12)
	gtx.Constraints.Max.Y = gtx.Dp(12)
	paint.FillShape(gtx.Ops, s.Color,
		clip.Ellipse{
			Max: gtx.Constraints.Max,
		}.Op(gtx.Ops))

	return layout.Dimensions{Size: gtx.Constraints.Max}
}
