package client

import (
	"fmt"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	"github.com/nbh-digital/goldchain/extensions/custodyfees/api"
	client "github.com/threefoldtech/rivine/pkg/client"
	types "github.com/threefoldtech/rivine/types"
)

// PluginClient is used to be able to get custody fee information
// for a coin output.
type PluginClient struct {
	client       client.BaseClient
	rootEndpoint string
}

// NewPluginConsensusClient creates a new PluginClient,
// that can be used for easy interaction with the Custody Fees Extension API exposed via the Consensus endpoints
func NewPluginConsensusClient(cli client.BaseClient) *PluginClient {
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
func NewPluginExplorerClient(cli client.BaseClient) *PluginClient {
	if cli == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		client:       cli,
		rootEndpoint: "/explorer",
	}
}

// GetCoinOutputInfo returns the custody fee related coin output information for a given coin output ID,
// returns an error only if the coin out never existed (spent or not).
func (cli *PluginClient) GetCoinOutputInfo(id types.CoinOutputID) (custodyfees.CoinOutputInfo, error) {
	var result api.CoinOutputInfoGet
	err := cli.client.HTTP().GetWithResponse(
		fmt.Sprintf("%s/custodyfees/coinoutput/%s?compute=true", cli.rootEndpoint, id.String()),
		&result)
	if err != nil {
		return custodyfees.CoinOutputInfo{}, fmt.Errorf(
			"failed to get custody fee info for coin output %s from daemon: %v", id.String(), err)
	}
	info := custodyfees.CoinOutputInfo{
		CreationTime:       result.CreationTime,
		CreationValue:      result.CreationValue,
		IsCustodyFee:       result.IsCustodyFee,
		Spent:              result.Spent,
		FeeComputationTime: result.FeeComputationTime,
	}
	if result.CustodyFee != nil {
		info.CustodyFee = *result.CustodyFee
	}
	if result.SpendableValue != nil {
		info.SpendableValue = *result.SpendableValue
	}
	return info, nil
}

// GetCoinOutputInfoOn returns the custody fee related coin output information for a given coin output ID,
// returns an error only if the coin out never existed (spent or not).
func (cli *PluginClient) GetCoinOutputInfoOn(id types.CoinOutputID, chainTime types.Timestamp) (custodyfees.CoinOutputInfo, error) {
	var result api.CoinOutputInfoGet
	err := cli.client.HTTP().GetWithResponse(
		fmt.Sprintf("%s/custodyfees/coinoutput/%s?compute=true&time=%d", cli.rootEndpoint, id.String(), chainTime),
		&result)
	if err != nil {
		return custodyfees.CoinOutputInfo{}, fmt.Errorf(
			"failed to get custody fee info for coin output %s from daemon: %v", id.String(), err)
	}
	info := custodyfees.CoinOutputInfo{
		CreationTime:       result.CreationTime,
		CreationValue:      result.CreationValue,
		IsCustodyFee:       result.IsCustodyFee,
		Spent:              result.Spent,
		FeeComputationTime: result.FeeComputationTime,
	}
	if result.CustodyFee != nil {
		info.CustodyFee = *result.CustodyFee
	}
	if result.SpendableValue != nil {
		info.SpendableValue = *result.SpendableValue
	}
	return info, nil
}

// GetCoinOutputInfoAt returns the custody fee related coin output information for a given coin output ID,
// returns an error only if the coin out never existed (spent or not).
func (cli *PluginClient) GetCoinOutputInfoAt(id types.CoinOutputID, chainHeight types.BlockHeight) (custodyfees.CoinOutputInfo, error) {
	var result api.CoinOutputInfoGet
	err := cli.client.HTTP().GetWithResponse(
		fmt.Sprintf("%s/custodyfees/coinoutput/%s?compute=true&height=%d", cli.rootEndpoint, id.String(), chainHeight),
		&result)
	if err != nil {
		return custodyfees.CoinOutputInfo{}, fmt.Errorf(
			"failed to get custody fee info for coin output %s from daemon: %v", id.String(), err)
	}
	info := custodyfees.CoinOutputInfo{
		CreationTime:       result.CreationTime,
		CreationValue:      result.CreationValue,
		IsCustodyFee:       result.IsCustodyFee,
		Spent:              result.Spent,
		FeeComputationTime: result.FeeComputationTime,
	}
	if result.CustodyFee != nil {
		info.CustodyFee = *result.CustodyFee
	}
	if result.SpendableValue != nil {
		info.SpendableValue = *result.SpendableValue
	}
	return info, nil
}

// GetCoinOutputInfoPreComputation returns the custody fee related coin output information for a given coin output ID,
// returns an error only if the coin out never existed (spent or not).
// Similar to `GetCoinOutputInfo` with the difference that the fee and spendable value aren't calculated yet.
func (cli *PluginClient) GetCoinOutputInfoPreComputation(id types.CoinOutputID) (custodyfees.CoinOutputInfoPreComputation, error) {
	var result api.CoinOutputInfoGet
	err := cli.client.HTTP().GetWithResponse(
		fmt.Sprintf("%s/custodyfees/coinoutput/%s?compute=false", cli.rootEndpoint, id.String()),
		&result)
	if err != nil {
		return custodyfees.CoinOutputInfoPreComputation{}, fmt.Errorf(
			"failed to get pre-compute custody fee info for coin output %s from daemon: %v", id.String(), err)
	}
	return custodyfees.CoinOutputInfoPreComputation{
		CreationTime:       result.CreationTime,
		CreationValue:      result.CreationValue,
		IsCustodyFee:       result.IsCustodyFee,
		Spent:              result.Spent,
		FeeComputationTime: result.FeeComputationTime,
	}, nil
}
