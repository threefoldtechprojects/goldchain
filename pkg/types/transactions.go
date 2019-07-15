package types

import "github.com/threefoldtech/rivine/types"

const (
	//MinterDefinitionTxVersion is the transaction version for the   minterdefinition transaction
	MinterDefinitionTxVersion types.TransactionVersion = iota + 128
	//CoinCreationTxVersion is the transaction version for the coin creation transaction
	CoinCreationTxVersion
	//CoinDestructionTxVersion is the transaction version for the coin destruction transaction
	CoinDestructionTxVersion
)

// Auth Coin Tx Extension Transaction Versions
const (
	TransactionVersionAuthAddressUpdateTx types.TransactionVersion = iota + 176
	TransactionVersionAuthConditionUpdateTx
)
