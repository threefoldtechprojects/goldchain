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

	"github.com/spf13/cobra"
)

func CreateConsensusSubCmds(ccli *rivinecli.CommandLineClient) error {
	bc, err := client.NewLazyBaseClientFromCommandLineClient(ccli)
	if err != nil {
		return err
	}

	consensusSubCmds := &consensusSubCmds{
		cli:      ccli,
		cfClient: NewPluginConsensusClient(bc),
	}

	// define commands
	var (
		getCoinOutputInfoCmd = &cobra.Command{
			Use:   "custodyfeeinfo id",
			Short: "Get all the custody-related info for a coin output",
			Run:   rivinecli.Wrap(consensusSubCmds.getCoinOutputInfo),
		}
	)

	// add commands as consensus sub commands
	ccli.ConsensusCmd.AddCommand(
		getCoinOutputInfoCmd,
	)

	// register flags
	getCoinOutputInfoCmd.Flags().Uint64Var(
		&consensusSubCmds.getCoinOutputInfoCfg.Timestamp, "time", 0,
		"look up the coin output info for a coin output, computing the fee for a specific timestamp")
	getCoinOutputInfoCmd.Flags().Uint64Var(
		&consensusSubCmds.getCoinOutputInfoCfg.Timestamp, "height", 0,
		"look up the coin output info for a coin output, computing the fee for a specific block height")
	getCoinOutputInfoCmd.Flags().BoolVar(
		&consensusSubCmds.getCoinOutputInfoCfg.ComputeFee, "fee", true,
		"do not compute the fee and spendable value as part of the result")
	getCoinOutputInfoCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &consensusSubCmds.getCoinOutputInfoCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))

	return nil
}

type consensusSubCmds struct {
	cli                  *rivinecli.CommandLineClient
	cfClient             *PluginClient
	getCoinOutputInfoCfg struct {
		Height       uint64
		Timestamp    uint64
		ComputeFee   bool
		EncodingType cli.EncodingType
	}
}

func (consensusSubCmds *consensusSubCmds) getCoinOutputInfo(str string) {
	var coid types.CoinOutputID
	err := coid.LoadString(str)
	if err != nil {
		cli.DieWithError("error while string-decoding coin output ID", err)
		return
	}
	var result interface{}
	if consensusSubCmds.getCoinOutputInfoCfg.ComputeFee {
		switch {
		case consensusSubCmds.getCoinOutputInfoCfg.Timestamp > 0:
			result, err = consensusSubCmds.cfClient.GetCoinOutputInfoOn(coid, types.Timestamp(consensusSubCmds.getCoinOutputInfoCfg.Timestamp))
		case consensusSubCmds.getCoinOutputInfoCfg.Height > 0:
			result, err = consensusSubCmds.cfClient.GetCoinOutputInfoAt(coid, types.BlockHeight(consensusSubCmds.getCoinOutputInfoCfg.Height))
		default:
			result, err = consensusSubCmds.cfClient.GetCoinOutputInfo(coid)
		}
	} else {
		result, err = consensusSubCmds.cfClient.GetCoinOutputInfoPreComputation(coid)
	}
	if err != nil {
		cli.DieWithError("error while getting coin output custody-related info from consensus", err)
		return
	}

	// encode depending on the encoding flag
	var encode func(interface{}) error
	switch consensusSubCmds.getCoinOutputInfoCfg.EncodingType {
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
