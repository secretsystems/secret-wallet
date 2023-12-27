package containers

import (
	"github.com/secretsystems/secret-wallet/containers/bottom_bar"
	"github.com/secretsystems/secret-wallet/containers/build_tx_modal"
	"github.com/secretsystems/secret-wallet/containers/confirm_modal"
	"github.com/secretsystems/secret-wallet/containers/image_modal"
	"github.com/secretsystems/secret-wallet/containers/listselect_modal"
	"github.com/secretsystems/secret-wallet/containers/node_status_bar"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/containers/password_modal"
	"github.com/secretsystems/secret-wallet/containers/prompt_modal"
	"github.com/secretsystems/secret-wallet/containers/qrcode_scan_modal"
	"github.com/secretsystems/secret-wallet/containers/recent_txs_modal"
)

func Load() {
	bottom_bar.LoadInstance()
	node_status_bar.LoadInstance()
	notification_modals.LoadInstance()
	recent_txs_modal.LoadInstance()
	build_tx_modal.LoadInstance()
	image_modal.LoadInstance()
	qrcode_scan_modal.LoadInstance()
	confirm_modal.LoadInstance()
	password_modal.LoadInstance()
	prompt_modal.LoadInstance()
	listselect_modal.LoadInstance()
}
