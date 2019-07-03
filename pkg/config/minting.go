package config

import (
	"github.com/threefoldtech/rivine/types"
)

// GetStandardGenesisMintCondition returns the standard network  minting condition
func GetStandardGenesisMintCondition() types.UnlockConditionProxy {
	// TODO: define final multisig condition
	address := ""
	var uh types.UnlockHash
	if err := uh.LoadString(address); err != nil {
		panic(err)
	}
	condition := types.NewCondition(types.NewUnlockHashCondition(uh))
	return condition
}

// GetTestnetGenesisMintCondition returns the testnet network  minting condition
func GetTestnetGenesisMintCondition() types.UnlockConditionProxy {
	// @leesmet
	address := "01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f"
	var uh types.UnlockHash
	if err := uh.LoadString(address); err != nil {
		panic(err)
	}
	condition := types.NewCondition(types.NewUnlockHashCondition(uh))
	return condition
}

// GetDevnetGenesisMintCondition returns the devnet network  minting condition
func GetDevnetGenesisMintCondition() types.UnlockConditionProxy {
	// belongs to wallet with mnemonic:
	// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
	address := "015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"
	var uh types.UnlockHash
	if err := uh.LoadString(address); err != nil {
		panic(err)
	}
	condition := types.NewCondition(types.NewUnlockHashCondition(uh))
	return condition
}
