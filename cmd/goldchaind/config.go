package main

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/daemon"
	"github.com/threefoldtech/rivine/types"
)

// ExtendedDaemonConfig contains all configurable variables for the deamon.
type ExtendedDaemonConfig struct {
	daemon.Config

	BootstrapPeers []modules.NetAddress

	GenesisMintCondition types.UnlockConditionProxy

	NetworkConfig daemon.NetworkConfig
}
