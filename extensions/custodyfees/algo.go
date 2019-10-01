package custodyfees

import (
	"math/big"

	"github.com/threefoldtech/rivine/types"
)

// AmountCustodyFeePairAfterXSeconds computes the value left over to spend after the, also returned,
// custody fee is subtracted from it. If only the value is required use `SpendableAmountAfterXSeconds` instead.
func AmountCustodyFeePairAfterXSeconds(c types.Currency, seconds types.Timestamp) (value, fee types.Currency) {
	value = SpendableAmountAfterXSeconds(c, seconds)
	fee = c.Sub(value)
	return
}

// SpendableAmountAfterXSeconds computes the spendable amount of value left over,
// after removing the custody fee to be paid for the given x seconds.
func SpendableAmountAfterXSeconds(c types.Currency, seconds types.Timestamp) types.Currency {
	// compute our duration tripplet, to keep the calculations small enough
	rd, rsh, rs := getDurationAsTripplet(seconds)

	// compute the ratios for each segment, allowing for 1, 2 or 3 ratios,
	// all merged together, to avoid rounding errors as much as possible
	var (
		nom = big.NewInt(1)
		// for nom the extra accuracy step is done at init. of x (as start of value)
		denom = new(big.Int).Mul(big.NewInt(1), extraAccuracyMultiplier)
	)
	multiplyRatio(rd, nom, denom, ratioDayNom, ratioDayDenom)
	multiplyRatio(rsh, nom, denom, ratioSemiHourNom, ratioSemiHourDenom)
	multiplyRatio(rs, nom, denom, ratioSecNom, ratioSecDenom)

	// keep our value as a more accurate amount, expressed as a big.Int
	x := new(big.Int).Mul(c.Big(), extraAccuracyMultiplier)

	// multiply our values with the total nominator
	x.Mul(x, nom)

	// divide our value now with the total denominator, rounding if needed,
	// by dividing by a total denominator we only round once
	_, r := x.QuoRem(x, denom, new(big.Int))
	if r.Cmp(new(big.Int).Div(denom, big.NewInt(2))) >= 0 {
		x.Add(x, big.NewInt(1))
	}

	// return final result as a currency amount
	return types.NewCurrency(x)
}

var (
	// ratio for seconds accuracy
	ratioSecDenom = big.NewInt(3456000000)
	ratioSecNom   = big.NewInt(3455999999)
	// ratio for semi-hour accuracy
	ratioSemiHourDenom = big.NewInt(1920000)
	ratioSemiHourNom   = big.NewInt(1919999)
	// ratio for day accuracy
	ratioDayDenom = big.NewInt(40000)
	ratioDayNom   = big.NewInt(39999)
)

var (
	extraAccuracyMultiplier = big.NewInt(1000)
)

func getDurationAsTripplet(seconds types.Timestamp) (rd, rsh, rs types.Timestamp) {
	rd = seconds / 86400
	seconds %= 86400
	rsh = seconds / 1800
	rs = seconds % 1800
	return
}

func multiplyRatio(power types.Timestamp, nom, denom, ratioNom, ratioDenom *big.Int) {
	if power == 0 {
		return
	}
	pow := big.NewInt(int64(power))
	nom.Mul(nom, new(big.Int).Exp(ratioNom, pow, nil))
	denom.Mul(denom, new(big.Int).Exp(ratioDenom, pow, nil))
}

type ratio struct {
	Nom   *big.Int
	Denom *big.Int
}
