package main

import (
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/daemon"

	"github.com/nbh-digital/goldchain/pkg/config"
	"github.com/nbh-digital/goldchain/pkg/types"
	authcointxcli "github.com/threefoldtech/rivine/extensions/authcointx/client"
	mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/client"
)

func main() {
	// create cli
	bchainInfo := config.GetBlockchainInfo()
	cliClient, err := NewCommandLineClient("", bchainInfo.Name, daemon.RivineUserAgent)
	if err != nil {
		panic(err)
	}

	// register goldchain-specific explorer commands
	mintingcli.CreateExploreCmd(cliClient.CommandLineClient)
	authcointxcli.CreateExploreAuthCoinInfoCmd(cliClient.CommandLineClient)

	// add cli wallet extension commands
	mintingcli.CreateWalletCmds(
		cliClient.CommandLineClient,
		types.MinterDefinitionTxVersion,
		types.CoinCreationTxVersion,
	)
	authcointxcli.CreateWalletCmds(
		cliClient.CommandLineClient,
		types.TransactionVersionAuthConditionUpdateTx,
		types.TransactionVersionAuthAddressUpdateTx,
	)

	// define preRun function
	cliClient.PreRunE = func(cfg *client.Config) (*client.Config, error) {
		if cfg == nil {
			bchainInfo := config.GetBlockchainInfo()
			chainConstants := config.GetStandardnetGenesis()
			daemonConstants := modules.NewDaemonConstants(bchainInfo, chainConstants)
			newCfg := client.ConfigFromDaemonConstants(daemonConstants)
			cfg = &newCfg
		}

		switch cfg.NetworkName {
		case config.NetworkNameStandard:
			RegisterStandardTransactions(cliClient.CommandLineClient)

			// overwrite standard network genesis block stamp,
			// as the genesis block is way earlier than the actual first block,
			// due to the hard reset at the bumpy/rough start
			cfg.GenesisBlockTimestamp = 1524168391 // timestamp of (standard) block #1

		case config.NetworkNameTest:
			RegisterTestnetTransactions(cliClient.CommandLineClient)

			// seems like testnet timestamp wasn't updated last time it was reset
			cfg.GenesisBlockTimestamp = 1522792547 // timestamp of (testnet) block #1

		case config.NetworkNameDev:
			RegisterDevnetTransactions(cliClient.CommandLineClient)

		default:
			return nil, fmt.Errorf("Network name %q not recognized", cfg.NetworkName)
		}

		return cfg, nil
	}

	// start cli
	if err := cliClient.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "client exited with an error: ", err)
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(cli.ExitCodeUsage)
	}
}
