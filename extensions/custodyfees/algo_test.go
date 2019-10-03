package custodyfees

import (
	"sync"
	"testing"

	"github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/pkg/config"
)

func TestAmountCustodyFeePairAfterXSeconds(t *testing.T) {
	testCases := []struct {
		InputValue     types.Currency
		Duration       types.Timestamp
		SpendableValue types.Currency
	}{
		{gft("0"), 0, gft("0")},
		{gft("0"), 500, gft("0")},
		{gft("1"), 1, gft("1")},
		{gft("1"), 50, gft("0.999999986")},
		{gft("0.000000001"), 999999, gft("0.000000001")},
		{gft("10"), 113, gft("9.999999673")},
		{gft("100"), 24 * 60 * 60, gft("99.9975")},
		{gft("40000"), 24 * 60 * 60, gft("39999")},
		{gft("500000000000"), 24 * 60 * 60, gft("499987500000")},
		{gft("500000000000"), 365 * 24 * 60 * 60, gft("495458196719.713017525")},
		{gft("35000.853"), 1, gft("35000.852989872")},
		{gft("35000.853"), 5404, gft("35000.798270685")},
		{gft("35000.853"), 13679330, gft("34862.586836388")},
		{gft("35000.853"), 157766400, gft("33438.96584286")},
		{gft("35000.853"), MaxCustodyFeeComputeDuration, gft("3.807056146")},
	}
	for i := 0; i < 5; i++ {
		var value, fee types.Currency
		for testIndex, testCase := range testCases {
			value, fee = AmountCustodyFeePairAfterXSeconds(testCase.InputValue, testCase.Duration)
			if value.Cmp(testCase.SpendableValue) != 0 {
				t.Errorf("run #%d: unexpected result in test case #%d: unexpected spendeable value: %s != %s", i+1, testIndex+1, gfts(value), gfts(testCase.SpendableValue))
				continue
			}
			expectedFee := testCase.InputValue.Sub(testCase.SpendableValue)
			if fee.Cmp(expectedFee) != 0 {
				t.Errorf("run #%d: unexpected result in test case #%d: unexpected custody fee: %s != %s", i+1, testIndex+1, gfts(fee), gfts(expectedFee))
			}
		}
	}
}

func BenchmarkAmountCustodyFeePairAfterXSeconds(b *testing.B) {
	var (
		c                 = gft("987432348584948439232921.493929483")
		d types.Timestamp = 157766400
	)
	for n := 0; n < b.N; n++ {
		AmountCustodyFeePairAfterXSeconds(c, d)
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
func gfts(c types.Currency) string {
	gftOnce.Do(gftOnceDo)
	return gftConvertor.ToCoinStringWithUnit(c)
}
