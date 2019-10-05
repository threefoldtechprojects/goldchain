package consensus

import (
	"encoding/hex"
	"sync"
	"testing"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/types"

	cftypes "github.com/nbh-digital/goldchain/extensions/custodyfees/types"
	"github.com/nbh-digital/goldchain/pkg/config"
)

func TestValidateCoinOutputsAreValid_ValidTxs(t *testing.T) {
	txs := []modules.ConsensusTransaction{
		nct(),
		nct(
			nco("0", ncfc(42)),
		),
		nct(
			nco("100", ncfc(1)),
		),
		nct(
			nco("0", ncfc(1)),
			nco("1", nnc()),
		),
		nct(
			nco("1", nnc()),
			nco("0", ncfc(1)),
		),
		nct(
			nco("1", nnc()),
			nco("3", ncfc(1)),
			nco("2", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
		),
		nct(
			nco("1", nnc()),
			nco("0", ncfc(1)),
			nco("2", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
			nco("4", nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
			nco("3", ntc(42, nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205"))),
			nco("5", nmsc(1,
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
		),
	}
	for idx, tx := range txs {
		err := ValidateCoinOutputsAreValid(tx, types.TransactionValidationContext{})
		if err != nil {
			t.Error(idx+1, err)
		}
	}
}

func TestValidateCoinOutputsAreValid_InvalidTxs(t *testing.T) {
	txs := []modules.ConsensusTransaction{
		nct(
			nco("0", nnc()),
		),
		nct(
			nco("0", nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
		),
		nct(
			nco("0", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
		),
		nct(
			nco("0", ntc(42, nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205"))),
		),
		nct(
			nco("0", nnc()),
			nco("1", nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
			nco("2", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
			nco("3", ntc(42, nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205"))),
		),
		nct(
			nco("3", nnc()),
			nco("0", nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
			nco("1", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
			nco("2", ntc(42, nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205"))),
		),
		nct(
			nco("2", nnc()),
			nco("3", nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
			nco("0", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
			nco("1", ntc(42, nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205"))),
		),
		nct(
			nco("1", nnc()),
			nco("2", nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205")),
			nco("3", nasc(
				"0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205",
				"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51",
				"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				42,
			)),
			nco("0", ntc(42, nuhc("0165c4d7cf3c52cab81fd7e82cd9e39d7fb8a1c7ab7515ac904299495244d0822c15841672f205"))),
		),
	}
	for idx, tx := range txs {
		err := ValidateCoinOutputsAreValid(tx, types.TransactionValidationContext{})
		if err == nil {
			t.Error(idx+1, "expected an error but none was received")
		}
	}
}

func nct(cos ...types.CoinOutput) modules.ConsensusTransaction {
	return modules.ConsensusTransaction{
		Transaction: types.Transaction{
			CoinOutputs: cos,
		},
	}
}

func nco(val string, condition types.MarshalableUnlockCondition) types.CoinOutput {
	return types.CoinOutput{
		Value:     gft(val),
		Condition: types.NewCondition(condition),
	}
}

func nnc() *types.NilCondition {
	return &types.NilCondition{}
}

func nuh(addr string) (uh types.UnlockHash) {
	err := uh.LoadString(addr)
	if err != nil {
		panic(err)
	}
	return
}

func nuhc(addr string) *types.UnlockHashCondition {
	return types.NewUnlockHashCondition(nuh(addr))
}

func nasc(sender, receiver, hashedSecret string, timeLock types.Timestamp) *types.AtomicSwapCondition {
	b, err := hex.DecodeString(hashedSecret)
	if err != nil {
		panic(err)
	}
	if len(b) != 32 {
		panic("invalid sized hashed secret, requires to be 32 bytes")
	}
	var hs types.AtomicSwapHashedSecret
	copy(hs[:], b[:])
	return &types.AtomicSwapCondition{
		Sender:       nuh(sender),
		Receiver:     nuh(receiver),
		TimeLock:     timeLock,
		HashedSecret: hs,
	}
}

func ntc(lockValue uint64, c types.MarshalableUnlockCondition) *types.TimeLockCondition {
	return types.NewTimeLockCondition(lockValue, c)
}

func nmsc(sigs uint64, addresses ...string) *types.MultiSignatureCondition {
	unlockHashes := make([]types.UnlockHash, 0, len(addresses))
	for _, addr := range addresses {
		unlockHashes = append(unlockHashes, nuh(addr))
	}
	if sigs == 0 {
		sigs = uint64(len(unlockHashes))
	}
	return types.NewMultiSignatureCondition(unlockHashes, sigs)
}

func ncfc(ts types.Timestamp) *cftypes.CustodyFeeCondition {
	return &cftypes.CustodyFeeCondition{
		ComputationTime: ts,
	}
}

var (
	gftOnce      sync.Once
	gftCfg       types.ChainConstants
	gftConvertor client.CurrencyConvertor
)

func gftOnceDo() {
	gftCfg = config.GetDevnetGenesis()
	gftConvertor = client.NewCurrencyConvertor(gftCfg.CurrencyUnits, config.TokenUnit)
}
func gft(val string) types.Currency {
	gftOnce.Do(gftOnceDo)
	c, err := gftConvertor.ParseCoinString(val)
	if err != nil {
		panic(err)
	}
	return c
}
