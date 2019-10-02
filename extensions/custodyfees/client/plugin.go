package client

import (
	"fmt"

	"github.com/nbh-digital/goldchain/extensions/custodyfees/api"
	client "github.com/threefoldtech/rivine/pkg/client"
	types "github.com/threefoldtech/rivine/types"
)

// PluginClient is used to be able to get custody fee information
// for a coin output.
type PluginClient struct {
	client       *client.BaseClient
	rootEndpoint string
}

// NewPluginConsensusClient creates a new PluginClient,
// that can be used for easy interaction with the Custody Fees Extension API exposed via the Consensus endpoints
func NewPluginConsensusClient(cli *client.BaseClient) *PluginClient {
	if cli == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		client:       cli,
		rootEndpoint: "/consensus",
	}
}

// NewPluginExplorerClient creates a new PluginClient,
// that can be used for easy interaction with the Custody Fees Extension API exposed via the Explorer endpoints
func NewPluginExplorerClient(cli *client.BaseClient) *PluginClient {
	if cli == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		client:       cli,
		rootEndpoint: "/explorer",
	}
}

func (cli *PluginClient) GetCoinOutputAge(id types.CoinOutputID) (types.Timestamp, error) {
	return cli.GetCoinOutputAgeOn(id, 0)
}
func (cli *PluginClient) GetCoinOutputAgeOn(id types.CoinOutputID, blockTime types.Timestamp) (types.Timestamp, error) {
	var result api.CoinOutputGetAge
	err := cli.client.HTTP().GetWithResponse(cli.rootEndpoint+"/custodyfees/coinoutput/age/"+id.String(), &result)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to get age for coin output %s from daemon: %v", id.String(), err)
	}
	return result.Age, nil
}

func (cli *PluginClient) GetCoinOutputValueCustodyFeePair(id types.CoinOutputID) (types.Currency, types.Currency, error) {
	return cli.GetCoinOutputValueCustodyFeePairOn(id, 0)
}
func (cli *PluginClient) GetCoinOutputValueCustodyFeePairOn(id types.CoinOutputID, blockTime types.Timestamp) (value, fee types.Currency, err error) {
	var result api.CoinOutputGetCustodyFee
	err = cli.client.HTTP().GetWithResponse(cli.rootEndpoint+"/custodyfees/coinoutput/fee/"+id.String(), &result)
	if err != nil {
		err = fmt.Errorf(
			"failed to get age for coin output %s from daemon: %v", id.String(), err)
		return
	}
	value = result.Value
	fee = result.Fee
	return
}
