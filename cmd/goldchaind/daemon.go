package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/nbh-digital/goldchain/pkg/config"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"
	goldchaintypes "github.com/nbh-digital/goldchain/pkg/types"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	authcointxapi "github.com/threefoldtech/rivine/extensions/authcointx/api"
	"github.com/threefoldtech/rivine/extensions/minting"
	mintingapi "github.com/threefoldtech/rivine/extensions/minting/api"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/blockcreator"
	"github.com/threefoldtech/rivine/modules/consensus"
	"github.com/threefoldtech/rivine/modules/explorer"
	"github.com/threefoldtech/rivine/modules/gateway"
	"github.com/threefoldtech/rivine/modules/transactionpool"
	"github.com/threefoldtech/rivine/modules/wallet"
	rivineapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/daemon"
)

func runDaemon(cfg ExtendedDaemonConfig, moduleIdentifiers daemon.ModuleIdentifierSet) error {
	// Print a startup message.
	fmt.Println("Loading...")
	loadStart := time.Now()

	var (
		i             int
		modulesToLoad = moduleIdentifiers.Len()
	)
	printModuleIsLoading := func(name string) {
		fmt.Printf("Loading %s (%d/%d)...\r\n", name, i, modulesToLoad)
		i++
	}

	// create our server already, this way we can fail early if the API addr is already bound
	fmt.Println("Binding API Address and serving the API...")
	srv, err := daemon.NewHTTPServer(cfg.APIaddr)
	if err != nil {
		return err
	}
	servErrs := make(chan error, 32)
	go func() {
		servErrs <- srv.Serve()
	}()

	ctx, cancel := context.WithCancel(context.Background())

	// load all modules

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// router to register all endpoints to
		router := httprouter.New()

		setupNetworkCfg, err := setupNetwork(cfg)
		if err != nil {
			servErrs <- fmt.Errorf("failed to create network config: %v", err)
			cancel()
			return
		}
		networkCfg := setupNetworkCfg.NetworkConfig
		err = networkCfg.Constants.Validate()
		if err != nil {
			servErrs <- fmt.Errorf("failed to validate network config: %v", err)
			cancel()
			return
		}

		// Initialize the Rivine modules
		var g modules.Gateway
		if moduleIdentifiers.Contains(daemon.GatewayModule.Identifier()) {
			printModuleIsLoading("gateway")
			g, err = gateway.New(cfg.RPCaddr, !cfg.NoBootstrap,
				filepath.Join(cfg.RootPersistentDir, modules.GatewayDir),
				cfg.BlockchainInfo, networkCfg.Constants, networkCfg.BootstrapPeers, cfg.VerboseLogging)
			if err != nil {
				servErrs <- err
				cancel()
				return
			}
			rivineapi.RegisterGatewayHTTPHandlers(router, g, cfg.APIPassword)
			defer func() {
				fmt.Println("Closing gateway...")
				err := g.Close()
				if err != nil {
					fmt.Println("Error during gateway shutdown:", err)
				}
			}()
		}

		var (
			cs modules.ConsensusSet

			// plugins
			authCoinTxPlugin *authcointx.Plugin
			mintingPlugin    *minting.Plugin
		)
		if moduleIdentifiers.Contains(daemon.ConsensusSetModule.Identifier()) {
			printModuleIsLoading("consensus set")
			cs, err = consensus.New(g, !cfg.NoBootstrap,
				filepath.Join(cfg.RootPersistentDir, modules.ConsensusDir),
				cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
			if err != nil {
				servErrs <- err
				cancel()
				return
			}
			rivineapi.RegisterConsensusHTTPHandlers(router, cs)
			defer func() {
				fmt.Println("Closing consensus set...")
				err := cs.Close()
				if err != nil {
					fmt.Println("Error during consensus set shutdown:", err)
				}
			}()

			// register the auth coin tx plugin
			// > NOTE: this also overwrites the standard tx controllers!!!!
			authCoinTxPlugin = authcointx.NewPlugin(
				setupNetworkCfg.GenesisAuthCondition,
				goldchaintypes.TransactionVersionAuthAddressUpdateTx,
				goldchaintypes.TransactionVersionAuthConditionUpdateTx,
				nil, // no custom opts
			)
			err = cs.RegisterPlugin(ctx, "authcointx", authCoinTxPlugin)
			if err != nil {
				servErrs <- fmt.Errorf("failed to register the auth coin tx extension: %v", err)
				err = authCoinTxPlugin.Close() //make sure any resources are released
				if err != nil {
					fmt.Println("Error during closing of the authCoinTxPlugin :", err)
				}
				cancel()
				return
			}
			// add the HTTP handlers for the auth coin tx extension as well
			authcointxapi.RegisterConsensusuthCoinHTTPHandlers(router, authCoinTxPlugin)

			// register the minting extension plugin
			mintingPlugin = minting.NewMintingPlugin(
				setupNetworkCfg.GenesisMintCondition,
				goldchaintypes.MinterDefinitionTxVersion,
				goldchaintypes.CoinCreationTxVersion,
				&minting.PluginOptions{
					CoinDestructionTransactionVersion: goldchaintypes.CoinDestructionTxVersion,
				},
			)
			err = cs.RegisterPlugin(ctx, "minting", mintingPlugin)
			if err != nil {
				servErrs <- fmt.Errorf("failed to register the minting extension: %v", err)
				err = mintingPlugin.Close() //make sure any resources are released
				if err != nil {
					fmt.Println("Error during closing of the mintingPlugin :", err)
				}
				cancel()
				return
			}
			// add the HTTP handlers for the auth coin tx extension as well
			mintingapi.RegisterConsensusMintingHTTPHandlers(router, mintingPlugin)
		}

		var tpool modules.TransactionPool
		if moduleIdentifiers.Contains(daemon.TransactionPoolModule.Identifier()) {
			printModuleIsLoading("transaction pool")
			tpool, err = transactionpool.New(cs, g,
				filepath.Join(cfg.RootPersistentDir, modules.TransactionPoolDir),
				cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
			if err != nil {
				servErrs <- err
				cancel()
				return
			}
			rivineapi.RegisterTransactionPoolHTTPHandlers(router, cs, tpool, cfg.APIPassword)
			defer func() {
				fmt.Println("Closing transaction pool...")
				err := tpool.Close()
				if err != nil {
					fmt.Println("Error during transaction pool shutdown:", err)
				}
			}()
		}
		var w modules.Wallet
		if moduleIdentifiers.Contains(daemon.WalletModule.Identifier()) {
			printModuleIsLoading("wallet")
			w, err = wallet.New(cs, tpool,
				filepath.Join(cfg.RootPersistentDir, modules.WalletDir),
				cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
			if err != nil {
				servErrs <- err
				cancel()
				return
			}
			rivineapi.RegisterWalletHTTPHandlers(router, w, cfg.APIPassword)
			defer func() {
				fmt.Println("Closing wallet...")
				err := w.Close()
				if err != nil {
					fmt.Println("Error during wallet shutdown:", err)
				}
			}()

		}
		var b modules.BlockCreator
		if moduleIdentifiers.Contains(daemon.BlockCreatorModule.Identifier()) {
			printModuleIsLoading("block creator")
			b, err = blockcreator.New(cs, tpool, w,
				filepath.Join(cfg.RootPersistentDir, modules.BlockCreatorDir),
				cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
			if err != nil {
				servErrs <- err
				cancel()
				return
			}
			// block creator has no API endpoints to register
			defer func() {
				fmt.Println("Closing block creator...")
				err := b.Close()
				if err != nil {
					fmt.Println("Error during block creator shutdown:", err)
				}
			}()
		}
		var e modules.Explorer
		if moduleIdentifiers.Contains(daemon.ExplorerModule.Identifier()) {
			printModuleIsLoading("explorer")
			e, err = explorer.New(cs,
				filepath.Join(cfg.RootPersistentDir, modules.ExplorerDir),
				cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
			if err != nil {
				servErrs <- err
				cancel()
				return
			}
			rivineapi.RegisterExplorerHTTPHandlers(router, cs, e, tpool)
			defer func() {
				fmt.Println("Closing explorer...")
				err := e.Close()
				if err != nil {
					fmt.Println("Error during explorer shutdown:", err)
				}
			}()

			// register extension HTTP handlers
			authcointxapi.RegisterExplorerAuthCoinHTTPHandlers(router, authCoinTxPlugin)
			mintingapi.RegisterExplorerMintingHTTPHandlers(router, mintingPlugin)
		}

		fmt.Println("Setting up root HTTP API handler...")

		// register our special daemon HTTP handlers
		router.GET("/daemon/constants", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			constants := modules.NewDaemonConstants(cfg.BlockchainInfo, networkCfg.Constants)
			rivineapi.WriteJSON(w, constants)
		})
		router.GET("/daemon/version", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			rivineapi.WriteJSON(w, daemon.Version{
				ChainVersion:    cfg.BlockchainInfo.ChainVersion,
				ProtocolVersion: cfg.BlockchainInfo.ProtocolVersion,
			})
		})
		router.POST("/daemon/stop", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			// can't write after we stop the server, so lie a bit.
			rivineapi.WriteSuccess(w)

			// need to flush the response before shutting down the server
			f, ok := w.(http.Flusher)
			if !ok {
				panic("Server does not support flushing")
			}
			f.Flush()

			if err := srv.Close(); err != nil {
				servErrs <- err
			}
			cancel()
		})

		// handle all our endpoints over a router,
		// which requires a user agent should one be configured
		srv.Handle("/", rivineapi.RequireUserAgentHandler(router, cfg.RequiredUserAgent))

		if cs != nil {
			cs.Start()
		}

		// Print a 'startup complete' message.
		startupTime := time.Since(loadStart)
		fmt.Println("Finished loading in", startupTime.Seconds(), "seconds")

		// wait until done
		<-ctx.Done()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	// wait for server to be killed or the process to be done
	select {
	case <-sigChan:
		fmt.Println("\rCaught stop signal, quitting...")
		srv.Close()
	case <-ctx.Done():
		fmt.Println("\rContext is done, quitting...")
	}

	cancel()
	wg.Wait()

	// return the first error which is returned
	return <-servErrs
}

type setupNetworkConfig struct {
	NetworkConfig        daemon.NetworkConfig
	GenesisMintCondition types.UnlockConditionProxy
	GenesisAuthCondition types.UnlockConditionProxy
}

// setupNetwork injects the correct chain constants and genesis nodes based on the chosen network,
// it also ensures that features added during the lifetime of the blockchain,
// only get activated on a certain block height, giving everyone sufficient time to upgrade should such features be introduced,
// it also creates the correct modules based on the given chain.
func setupNetwork(cfg ExtendedDaemonConfig) (setupNetworkConfig, error) {
	// return the network configuration, based on the network name,
	// which includes the genesis block as well as the bootstrap peers
	switch cfg.BlockchainInfo.NetworkName {
	case config.NetworkNameStandard:
		return setupNetworkConfig{}, errors.New("standard net is disabled for goldchain, it is not ready for production") // TODO: enable again

		// constants := config.GetStandardnetGenesis()
		// networkConfig := config.GetStandardDaemonNetworkConfig()
		// genesisAuthCondition := config.GetStandardnetGenesisAuthCoinCondition()

		// // Register the transaction controllers for all transaction versions
		// // supported on the standard network
		// goldchaintypes.RegisterTransactionTypesForStandardNetwork(constants.CurrencyUnits.OneCoin, networkConfig)

		// todo set bootstrap peers only if not set yet
		// cfg.BootstrapPeers = config.GetStandardnetBootstrapPeers()

		// // return the standard genesis block and bootstrap peers
		// return setupNetworkConfig{
		// 	NetworkConfig: daemon.NetworkConfig{
		// 		Constants:      constants,
		// 		BootstrapPeers: cfg.BootstrapPeers,
		// 	},
		// 	GenesisAuthCondition: genesisAuthCondition,
		// }, nil

	case config.NetworkNameTest:

		constants := config.GetTestnetGenesis()
		genesisMintCondition := config.GetTestnetGenesisMintCondition()
		genesisAuthCondition := config.GetTestnetGenesisAuthCoinCondition()

		bootstrapPeers := cfg.BootstrapPeers
		if len(bootstrapPeers) == 0 {
			bootstrapPeers = config.GetTestnetBootstrapPeers()
		}

		// return the testnet genesis block and bootstrap peers
		return setupNetworkConfig{
			NetworkConfig: daemon.NetworkConfig{
				Constants:      constants,
				BootstrapPeers: bootstrapPeers,
			},
			GenesisMintCondition: genesisMintCondition,
			GenesisAuthCondition: genesisAuthCondition,
		}, nil

	case config.NetworkNameDev:

		constants := config.GetDevnetGenesis()
		genesisMintCondition := config.GetDevnetGenesisMintCondition()
		genesisAuthCondition := config.GetDevnetGenesisAuthCoinCondition()

		bootstrapPeers := cfg.BootstrapPeers
		if len(bootstrapPeers) == 0 {
			bootstrapPeers = config.GetDevnetBootstrapPeers()
		}

		// return the devnet genesis block and bootstrap peers
		return setupNetworkConfig{
			NetworkConfig: daemon.NetworkConfig{
				Constants:      constants,
				BootstrapPeers: bootstrapPeers,
			},
			GenesisMintCondition: genesisMintCondition,
			GenesisAuthCondition: genesisAuthCondition,
		}, nil

	default:
		// network isn't recognised
		return setupNetworkConfig{}, fmt.Errorf(
			"Netork name %q not recognized", cfg.BlockchainInfo.NetworkName)
	}
}
