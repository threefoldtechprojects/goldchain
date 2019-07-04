package main

import (
	"github.com/threefoldtech/rivine/pkg/daemon"
)

// ExtendedDaemonConfig contains all configurable variables for the deamon.
type ExtendedDaemonConfig struct {
	daemon.Config
}

// DefaultConfig returns the default daemon configuration
func DefaultConfig() daemon.Config {
	cfg := daemon.DefaultConfig()
	cfg.APIaddr = "localhost:22112"
	cfg.RPCaddr = ":22110"
	return cfg
}
