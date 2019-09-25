package main

import (
	goldchaintypes "github.com/nbh-digital/goldchain/pkg/types"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	authcointxcli "github.com/threefoldtech/rivine/extensions/authcointx/client"
	"github.com/threefoldtech/rivine/extensions/minting"
	mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/pkg/client"
)

func RegisterDevnetTransactions(bc *client.BaseClient) {
	registerTransactions(bc)
}

func RegisterTestnetTransactions(bc *client.BaseClient) {
	registerTransactions(bc)
}

func registerTransactions(bc *client.BaseClient) {
	// create minting plugin client...
	mintingCLI := mintingcli.NewPluginConsensusClient(bc)
	// ...and register minting types
	types.RegisterTransactionVersion(goldchaintypes.TransactionVersionMinterDefinition, minting.MinterDefinitionTransactionController{
		MintConditionGetter: mintingCLI,
		TransactionVersion:  goldchaintypes.TransactionVersionMinterDefinition,
	})
	types.RegisterTransactionVersion(goldchaintypes.TransactionVersionCoinCreation, minting.CoinCreationTransactionController{
		MintConditionGetter: mintingCLI,
		TransactionVersion:  goldchaintypes.TransactionVersionCoinCreation,
	})
	types.RegisterTransactionVersion(goldchaintypes.TransactionVersionCoinDestruction, minting.CoinDestructionTransactionController{
		TransactionVersion: goldchaintypes.TransactionVersionCoinDestruction,
	})

	// create coin auth tx plugin client...
	authCoinTxCLI := authcointxcli.NewPluginConsensusClient(bc)
	// ...and register coin auth tx types
	types.RegisterTransactionVersion(goldchaintypes.TransactionVersionAuthConditionUpdate, authcointx.AuthConditionUpdateTransactionController{
		AuthInfoGetter:     authCoinTxCLI,
		TransactionVersion: goldchaintypes.TransactionVersionAuthConditionUpdate,
	})
	types.RegisterTransactionVersion(goldchaintypes.TransactionVersionAuthAddressUpdate, authcointx.AuthAddressUpdateTransactionController{
		AuthInfoGetter:     authCoinTxCLI,
		TransactionVersion: goldchaintypes.TransactionVersionAuthAddressUpdate,
	})
}
