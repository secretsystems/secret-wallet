package node_manager

import (
	"github.com/deroproject/derohe/walletapi"
	"github.com/secretsystems/secret-wallet/app_db"

	"github.com/secretsystems/secret-wallet/settings"
)

var CurrentNode *app_db.NodeConnection

func Load() error {
	endpoint := settings.App.NodeEndpoint
	if endpoint != "" {
		var nodeConn *app_db.NodeConnection

		conn, err := app_db.GetNodeConnectionByEndpoint(endpoint)
		if err != nil {
			return err
		}

		nodeConn = &conn
		if nodeConn == nil {
			nodeConn = &app_db.NodeConnection{
				Name:     "",
				Endpoint: endpoint,
			}
		}

		err = Connect(*nodeConn, false)
		if err != nil {
			return err
		}

		CurrentNode = nodeConn
	}

	return nil
}

func Connect(nodeConn app_db.NodeConnection, save bool) error {
	endpoint := nodeConn.Endpoint

	err := walletapi.Connect(endpoint)
	if err != nil {
		return err
	}

	CurrentNode = &nodeConn
	settings.App.NodeEndpoint = nodeConn.Endpoint

	if save {
		err := settings.Save()
		if err != nil {
			return err
		}
	}

	return nil
}
