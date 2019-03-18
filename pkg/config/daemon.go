package config

import (
	"github.com/threefoldtech/rivine/types"
)

// DaemonNetworkConfig defines network-specific tfchain constants.
type DaemonNetworkConfig struct {
	FoundationPoolAddress types.UnlockHash
}

// GetStandardDaemonNetworkConfig returns the standard network config for the daemon
func GetStandardDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		// TODO: define final address
		FoundationPoolAddress: unlockHashFromHex(""),
	}
}

// GetTestnetDaemonNetworkConfig returns the testnet network config for the daemon
func GetTestnetDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		// TODO: define final address
		FoundationPoolAddress: unlockHashFromHex(""),
	}
}

// GetDevnetDaemonNetworkConfig returns the devnet network config for the daemon
func GetDevnetDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		// belongs to wallet with mnemonic:
		// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
		FoundationPoolAddress: unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"),
	}
}
