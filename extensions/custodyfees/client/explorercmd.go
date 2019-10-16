package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/client"
	rivinecli "github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees/api"

	"github.com/spf13/cobra"
)

func CreateExplorerSubCmds(ccli *rivinecli.CommandLineClient) error {
	bc, err := client.NewLazyBaseClientFromCommandLineClient(ccli)
	if err != nil {
		return err
	}

	explorerSubCmds := &explorerSubCmds{
		cli:      ccli,
		cfClient: NewPluginExplorerClient(bc),
	}

	// define commands
	var (
		getCoinOutputInfoCmd = &cobra.Command{
			Use:   "custodyfeeinfo id",
			Short: "Get all the custody-related info for a coin output",
			Run:   rivinecli.Wrap(explorerSubCmds.getCoinOutputInfo),
		}
		getChainFactsCmd = &cobra.Command{
			Use:   "chainfacts",
			Short: "Get the latest Chain Facts",
			Run:   rivinecli.Wrap(explorerSubCmds.getChainFacts),
		}
	)

	// add commands as explorer sub commands
	ccli.ExploreCmd.AddCommand(
		getCoinOutputInfoCmd,
		getChainFactsCmd,
	)

	// register flags
	getCoinOutputInfoCmd.Flags().Uint64Var(
		&explorerSubCmds.getCoinOutputInfoCfg.Timestamp, "time", 0,
		"look up the coin output info for a coin output, computing the fee for a specific timestamp")
	getCoinOutputInfoCmd.Flags().Uint64Var(
		&explorerSubCmds.getCoinOutputInfoCfg.Timestamp, "height", 0,
		"look up the coin output info for a coin output, computing the fee for a specific block height")
	getCoinOutputInfoCmd.Flags().BoolVar(
		&explorerSubCmds.getCoinOutputInfoCfg.ComputeFee, "fee", true,
		"do not compute the fee and spendable value as part of the result")
	getCoinOutputInfoCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &explorerSubCmds.getCoinOutputInfoCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))
	getChainFactsCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &explorerSubCmds.getChainFactsCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))

	return nil
}

type explorerSubCmds struct {
	cli                  *rivinecli.CommandLineClient
	cfClient             *PluginClient
	getCoinOutputInfoCfg struct {
		Height       uint64
		Timestamp    uint64
		ComputeFee   bool
		EncodingType cli.EncodingType
	}
	getChainFactsCfg struct {
		EncodingType cli.EncodingType
	}
}

func (explorerSubCmds *explorerSubCmds) getCoinOutputInfo(str string) {
	var coid types.CoinOutputID
	err := coid.LoadString(str)
	if err != nil {
		cli.DieWithError("error while string-decoding coin output ID", err)
		return
	}
	var result interface{}
	if explorerSubCmds.getCoinOutputInfoCfg.ComputeFee {
		switch {
		case explorerSubCmds.getCoinOutputInfoCfg.Timestamp > 0:
			result, err = explorerSubCmds.cfClient.GetCoinOutputInfoOn(coid, types.Timestamp(explorerSubCmds.getCoinOutputInfoCfg.Timestamp))
		case explorerSubCmds.getCoinOutputInfoCfg.Height > 0:
			result, err = explorerSubCmds.cfClient.GetCoinOutputInfoAt(coid, types.BlockHeight(explorerSubCmds.getCoinOutputInfoCfg.Height))
		default:
			result, err = explorerSubCmds.cfClient.GetCoinOutputInfo(coid)
		}
	} else {
		result, err = explorerSubCmds.cfClient.GetCoinOutputInfoPreComputation(coid)
	}
	if err != nil {
		cli.DieWithError("error while getting coin output custody-related info from explorer", err)
		return
	}

	// encode depending on the encoding flag
	var encode func(interface{}) error
	switch explorerSubCmds.getCoinOutputInfoCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b, err := rivbin.Marshal(v)
			if err != nil {
				return err
			}
			fmt.Println(hex.EncodeToString(b))
			return nil
		}
	}
	err = encode(result)
	if err != nil {
		cli.DieWithError("failed to encode coin output info", err)
	}
}

func (explorerSubCmds *explorerSubCmds) getChainFacts() {
	var result api.ChainFactsGet
	err := explorerSubCmds.cli.GetWithResponse("/explorer/custodyfees/metrics/chain", &result)
	if err != nil {
		cli.DieWithError("failed get chain facts info", err)
		return
	}

	// encode depending on the encoding flag
	var encode func(interface{}) error
	switch explorerSubCmds.getCoinOutputInfoCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b, err := rivbin.Marshal(v)
			if err != nil {
				return err
			}
			fmt.Println(hex.EncodeToString(b))
			return nil
		}
	}
	err = encode(result)
	if err != nil {
		cli.DieWithError("failed to encode coin output info", err)
	}
}
