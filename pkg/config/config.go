package config

import (
	"math/big"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/types"
	"github.com/threefoldtech/rivine/modules"
)

var (
	rawVersion = "v0.2"
	// Version of the chain binaries.
	//
	// Value is defined by a private build flag,
	// or hardcoded to the latest released tag as fallback.
	Version build.ProtocolVersion
)

const (
	// TokenUnit defines the unit of one Token.
	TokenUnit = "GFT"
	// TokenChainName defines the name of the chain.
	TokenChainName = "goldchain"
)

// chain network names
const (
	
	NetworkNameDevnet = "devnet"
	
	NetworkNameTestnet = "testnet"
	
)

func GetDefaultGenesis() types.ChainConstants {
	return GetTestnetGenesis()
}

// GetBlockchainInfo returns the naming and versioning of tfchain.
func GetBlockchainInfo() types.BlockchainInfo {
	return types.BlockchainInfo{
		Name:            TokenChainName,
		NetworkName:     NetworkNameTestnet,
		CoinUnit:        TokenUnit,
		ChainVersion:    Version,       // use our own blockChain/build version
		ProtocolVersion: build.Version, // use latest available rivine protocol version
	}
}

func GetDevnetGenesis() types.ChainConstants {
	cfg := types.DevnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersion(1)
	cfg.GenesisTransactionVersion = types.TransactionVersion(1)

	// size limits
	cfg.BlockSizeLimit = 2000000
	cfg.ArbitraryDataSizeLimit = 83

	// block time
	cfg.BlockFrequency = 12

	// Time to MaturityDelay
	cfg.MaturityDelay = 10

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1519200000)

	cfg.MedianTimestampWindow = 11

	// block window for difficulty
	cfg.TargetWindow = 20

	cfg.MaxAdjustmentUp = big.NewRat(120, 100)
	cfg.MaxAdjustmentDown = big.NewRat(100, 120)

	cfg.FutureThreshold = 120
	cfg.ExtremeFutureThreshold = 240

	cfg.StakeModifierDelay = 2000

	// Time it takes before transferred blockstakes can be used
	cfg.BlockStakeAging = 64

	// Coins you receive when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(10)// Minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Mul64(1)
	cfg.TransactionFeeCondition = types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
	

	// Set Transaction Pool config
	cfg.TransactionPool = types.TransactionPoolConstants{
		TransactionSizeLimit:    16000,
		TransactionSetSizeLimit: 250000,
		PoolSizeLimit:           20000000,
	}

	// allocate initial coin outputs
	cfg.GenesisCoinDistribution = []types.CoinOutput{ 
		{
			Value: cfg.CurrencyUnits.OneCoin.Mul64(100000000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	// allocate initial block stake outputs
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{ 
		{
		Value:     types.NewCurrency64(3000),
		Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	return cfg
}

func GetDevnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{ 
		"localhost:23112",
	}
}

func GetDevnetGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))}

func GetDevnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))}

func GetTestnetGenesis() types.ChainConstants {
	cfg := types.TestnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersion(1)
	cfg.GenesisTransactionVersion = types.TransactionVersion(1)

	// size limits
	cfg.BlockSizeLimit = 2000000
	cfg.ArbitraryDataSizeLimit = 83

	// block time
	cfg.BlockFrequency = 120

	// Time to MaturityDelay
	cfg.MaturityDelay = 720

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1564142400)

	cfg.MedianTimestampWindow = 11

	// block window for difficulty
	cfg.TargetWindow = 1000

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 3600
	cfg.ExtremeFutureThreshold = 7200

	cfg.StakeModifierDelay = 2000

	// Time it takes before transferred blockstakes can be used
	cfg.BlockStakeAging = 64

	// Coins you receive when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(0)// Minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Mul64(001).Div64(1000)

	// Set Transaction Pool config
	cfg.TransactionPool = types.TransactionPoolConstants{
		TransactionSizeLimit:    16000,
		TransactionSetSizeLimit: 250000,
		PoolSizeLimit:           20000000,
	}

	// allocate initial coin outputs
	cfg.GenesisCoinDistribution = []types.CoinOutput{ 
		{
			Value: cfg.CurrencyUnits.OneCoin.Mul64(100000000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f"))),
		},
	}

	// allocate initial block stake outputs
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{ 
		{
		Value:     types.NewCurrency64(3000),
		Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f"))),
		},
	}

	return cfg
}

func GetTestnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{ 
		"bootstrap1.testnet.nbh-digital.com:22112",
		"bootstrap2.testnet.nbh-digital.com:22112",
		"bootstrap3.testnet.nbh-digital.com:22112",
		"bootstrap4.testnet.nbh-digital.com:22112",
		"bootstrap5.testnet.nbh-digital.com:22112",
	}
}

func GetTestnetGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f")))}

func GetTestnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f")))}


func init() {
	Version = build.MustParse(rawVersion)
}
