package main

import (
	"github.com/nbh-digital/goldchain/pkg/types"
	"github.com/threefoldtech/rivine/extensions/minting"
	"github.com/threefoldtech/rivine/modules"
	rivinetypes "github.com/threefoldtech/rivine/types"
)

func registerMintingExtension(cs modules.ConsensusSet, genesisMintCondition rivinetypes.UnlockConditionProxy) (err error) {
	plugin := minting.NewMintingPlugin(genesisMintCondition, types.MinterDefinitionTxVersion, types.CoinCreationTxVersion)
	//TODO replace with a decent context once this is implemented in rivine
	cancel := make(chan struct{})
	err = cs.RegisterPlugin("minting", plugin, cancel)
	return
}
