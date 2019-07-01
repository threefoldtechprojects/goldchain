package types

const (
	//MinterDefinitionTxVersion is the transaction version for the   minterdefinition transaction
	MinterDefinitionTxVersion = iota + 128
	//CoinCreationTxVersion is the transaction version for the coin creation transaction
	CoinCreationTxVersion
)

// Auth Coin Tx Extension Transaction Versions
const (
	TransactionVersionAuthAddressUpdateTx = iota + 176
	TransactionVersionAuthConditionUpdateTx
)
