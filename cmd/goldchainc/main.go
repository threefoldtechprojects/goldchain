package main

import (
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/daemon"

	"github.com/nbh-digital/goldchain/pkg/config"

	cfcli "github.com/nbh-digital/goldchain/extensions/custodyfees/client"
	gccli "github.com/nbh-digital/goldchain/pkg/client"
	"github.com/nbh-digital/goldchain/pkg/types"
	authcointxcli "github.com/threefoldtech/rivine/extensions/authcointx/client"
	mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/client"
)

func main() {
	// create cli
	bchainInfo := config.GetBlockchainInfo()
	cliClient, err := NewCommandLineClient("http://localhost:22110", bchainInfo.Name, daemon.RivineUserAgent)
	exitIfError(err)

	// register goldchain-specific explorer commands
	err = mintingcli.CreateExploreCmd(cliClient.CommandLineClient)
	exitIfError(err)
	err = cfcli.CreateExplorerSubCmds(cliClient.CommandLineClient)
	exitIfError(err)
	err = mintingcli.CreateConsensusCmd(cliClient.CommandLineClient)
	exitIfError(err)
	err = cfcli.CreateConsensusSubCmds(cliClient.CommandLineClient)
	exitIfError(err)

	// add cli wallet extension commands
	err = gccli.CreateMintingWalletCmds(cliClient.CommandLineClient)
	exitIfError(err)

	err = authcointxcli.CreateConsensusAuthCoinInfoCmd(cliClient.CommandLineClient)
	exitIfError(err)
	err = authcointxcli.CreateExploreAuthCoinInfoCmd(cliClient.CommandLineClient)
	exitIfError(err)
	authcointxcli.CreateWalletCmds(
		cliClient.CommandLineClient,
		types.TransactionVersionAuthConditionUpdate,
		types.TransactionVersionAuthAddressUpdate,
	)

	// define preRun function
	cliClient.PreRunE = func(cfg *client.Config) (*client.Config, error) {
		if cfg == nil {
			bchainInfo := config.GetBlockchainInfo()
			chainConstants := config.GetDefaultGenesis()
			daemonConstants := modules.NewDaemonConstants(bchainInfo, chainConstants, nil)
			newCfg := client.ConfigFromDaemonConstants(daemonConstants)
			cfg = &newCfg
		}

		bc, err := client.NewLazyBaseClientFromCommandLineClient(cliClient.CommandLineClient)
		if err != nil {
			return nil, err
		}

		switch cfg.NetworkName {

		case config.NetworkNameDevnet:
			RegisterDevnetTransactions(bc)
			cfg.GenesisBlockTimestamp = 1519200000 // timestamp of block #1

		case config.NetworkNameTestnet:
			RegisterTestnetTransactions(bc)
			cfg.GenesisBlockTimestamp = 1564142400 // timestamp of block #1

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

func exitIfError(err error) {
	if err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "client exited during setup with an error:", err)
	os.Exit(cli.ExitCodeGeneral)
}
