package config

import (
	"github.com/threefoldtech/rivine/types"
)

// DaemonNetworkConfig defines network-specific constants.
type DaemonNetworkConfig struct {
	FoundationPoolAddress types.UnlockHash
}

func GetDevnetDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		FoundationPoolAddress: unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"),
	}
}

func GetTestnetDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{}
}
