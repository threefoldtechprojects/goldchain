package main

import (
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	"github.com/nbh-digital/goldchain/pkg/config"
)

/*
If we assume a daily fee of `0.0025%` of the amount of tokens that become unspendable of the total value of a coin output,
than we could also express that as a ratio of `1/40000`, such that `39999/40000` remain spendable.

This results in a Geometric sequence of `f(x) = S * (39999/40000)^x`, where S is the start value of a created output,
allowing you to compute the amount of tokens spendable at the `x`th day after creation.

On a blockchain level we work however with a granularity of seconds and thus we need to divide this ratio by `86 400`, the amount of seconds in one day,
giving us a ratio of `1 / 3 456 000 000`, resulting in a geometric sequence of `f(x) = S * (3 455 999 999 / 3 456 000 000)^n`, allowing you to compute the amount
of tokens that are spendable after `x` seconds.

> (!) Following the standard practice of the Financial Industry, we'll round amounts that have a precision greater then 9,
> the maximum precision allowed for GFT on the GoldChain at the time of writing this.

Now that we have these definitions, let's see some examples:

| start amount | `1s` | `5404s` (3 semi-hour(s), 4 second(s)) | `13679330s` (158 day(s), 15 semi-hour(s), 1130 second(s)) | `157766400s` (1826 day(s), 0 semi-hour(s), 0 second(s)) |
| - | - | - | - | - |
| `0.000000001 GFT` | `0.000000001 GFT` | `0.000000001 GFT` | `0.000000001 GFT` | `0.000000001 GFT` |
| `0.00000001 GFT` | `0.00000001 GFT` | `0.00000001 GFT` | `0.00000001 GFT` | `0.00000001 GFT` |
| `0.0000001 GFT` | `0.0000001 GFT` | `0.0000001 GFT` | `0.0000001 GFT` | `0.000000096 GFT` |
| `0.000001 GFT` | `0.000001 GFT` | `0.000001 GFT` | `0.000000996 GFT` | `0.000000955 GFT` |
| `0.0015 GFT` | `0.0015 GFT` | `0.001499998 GFT` | `0.001494074 GFT` | `0.001433064 GFT` |
| `1 GFT` | `1 GFT` | `0.999998436 GFT` | `0.996049634 GFT` | `0.95537574 GFT` |
| `35000.853 GFT` | `35000.852989872 GFT` | `35000.798270685 GFT` | `34862.586836388 GFT` | `33438.96584286 GFT` |
| `50000000 GFT` | `49999999.985532407 GFT` | `49999921.81717041 GFT` | `49802481.722928799 GFT` | `47768787.010505162 GFT` |

> (!) to stay even more accurate we multiply the amount by 1000 to have more
> precision than the chain allows, only at the end do we divide back by that precision to return to our original precision.

As you can see, when the amounts are too small no fee is deducted due to rounding.
Knowing that 1 GFT equals 1 Gram of gold, these amounts (<= `0.0000001 GFT`) are way too small to really are about this
side effect. On top of that it is not beneficial to split up all your coin outputs into such small coin outputs,
given that the benefit would immediately turn on you by all the miner fees you would have to pay for each of those
split transactions.

By rounding in a deterministic manner, and having the value registered on which it is calculated,
every consensus node can recompute the required custody fee and know up to lowest precision if that custody fee is correctly paid out.

----

A computational note: while we would want to completely compute on a granularity of a second,
it is computational not possible. Therefore we transform the age of a spent coin output from seconds
to a triplet of days, semi-hours and seconds. Let's see what that means with an example.

First of all, let's take our original function, where S is the value of the coin output,
giving the amount of spendable coins after x seconds:

```
f(x) = S * (3 455 999 999 / 3 456 000 000)^n
```

In our example let's define n as `181 905 seconds`, which equals to exactly `2 days, 5 semi hours and 105 seconds`,
and thus we get:

```
f(x) = S * (3 455 999 999 / 3 456 000 000)^181905
	 = S * (3 455 999 999 / 3 456 000 000)^(2 d) * (3 455 999 999 / 3 456 000 000)^(2 1/2h)
		 * (3 455 999 999 / 3 456 000 000)^105
	 ≃ S * (39999 / 40000)^2 * (1 919 999 / 1 920 000)^5
	     * (3 455 999 999 / 3 456 000 000)^105
```

Which is not as accurate as remaining in seconds, but it is accurate enough, that it shouldn't matter.
On top of that it allows us to actually compute, as computing a power of 181905 is not going to happen.
*/

var (
	gftOnce      sync.Once
	gftCfg       types.ChainConstants
	gftConvertor client.CurrencyConvertor
)

func gftOnceDo() {
	gftCfg = config.GetTestnetGenesis()
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

var (
	_durations = []int64{
		1,
		5404,
		13679330,
		157766400,
	}

	_amounts = []types.Currency{
		gft("0.000000001"),
		gft("0.00000001"),
		gft("0.0000001"),
		gft("0.000001"),
		gft("0.0015"),
		gft("1"),
		gft("35000.853"),
		gft("50000000"),
	}
)

func getDurationAsTripplet(seconds types.Timestamp) (rd, rsh, rs types.Timestamp) {
	rd = seconds / 86400
	seconds %= 86400
	rsh = seconds / 1800
	rs = seconds % 1800
	return
}

func main() {
	fmt.Printf("| start amount |")
	for _, duration := range _durations {
		rd, rsh, rs := getDurationAsTripplet(types.Timestamp(duration))
		if rd > 0 {
			fmt.Printf(" `%ds` (%d day(s), %d semi-hour(s), %d second(s)) |", duration, rd, rsh, rs)
		} else if rsh > 0 {
			fmt.Printf(" `%ds` (%d semi-hour(s), %d second(s)) |", duration, rsh, rs)
		} else {
			fmt.Printf(" `%ds` |", duration)
		}
	}
	fmt.Printf("\n| - |")
	for range _durations {
		fmt.Printf(" - |")
	}
	fmt.Println()
	for _, amount := range _amounts {
		fmt.Printf("| `%s` |", gfts(amount))
		for _, duration := range _durations {
			spendable := custodyfees.SpendableAmountAfterXSeconds(amount, types.Timestamp(duration))
			fmt.Printf(" `%s` |", gfts(spendable))
		}
		fmt.Println()
	}
}
