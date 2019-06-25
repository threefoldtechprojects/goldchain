package types

import (
	"github.com/nbh-digital/goldchain/pkg/config"

	"github.com/threefoldtech/rivine/types"
)

// RegisterTransactionTypesForStandardNetwork registers he transaction controllers
// for all transaction versions supported on the standard network.
func RegisterTransactionTypesForStandardNetwork(oneCoin types.Currency, cfg config.DaemonNetworkConfig) {
	//Just stick to rivine defaults.
}

// RegisterTransactionTypesForTestNetwork registers he transaction controllers
// for all transaction versions supported on the test network.
func RegisterTransactionTypesForTestNetwork(oneCoin types.Currency, cfg config.DaemonNetworkConfig) {
	//Just stick to rivine defaults.

}

// RegisterTransactionTypesForDevNetwork registers he transaction controllers
// for all transaction versions supported on the dev network.
func RegisterTransactionTypesForDevNetwork(oneCoin types.Currency, cfg config.DaemonNetworkConfig) {
	//Just stick to rivine defaults.
}
