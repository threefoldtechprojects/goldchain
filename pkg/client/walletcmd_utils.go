package client

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/threefoldtech/rivine/types"
)

type outputPair struct {
	Condition types.UnlockConditionProxy
	Value     types.Currency
}

// parseCurrencyString takes the string representation of a currency value
type parseCurrencyString func(string) (types.Currency, error)

func stringToBlockStakes(input string) (types.Currency, error) {
	bsv, err := strconv.ParseUint(input, 10, 64)
	return types.NewCurrency64(bsv), err
}

func parsePairedOutputs(args []string, parseCurrency parseCurrencyString) (pairs []outputPair, err error) {
	argn := len(args)
	if argn < 2 {
		err = errors.New("not enough arguments, at least 2 required")
		return
	}
	if argn%2 != 0 {
		err = errors.New("arguments have to be given in pairs of '<dest>|<rawCondition>'+'<value>'")
		return
	}

	for i := 0; i < argn; i += 2 {
		// parse value first, as it's the one without any possibility of ambiguity
		var pair outputPair
		pair.Value, err = parseCurrency(args[i+1])
		if err != nil {
			err = fmt.Errorf("failed to parse amount/value for output #%d: %v", i/2, err)
			return
		}

		// try to parse it as an unlock hash
		var uh types.UnlockHash
		err = uh.LoadString(args[i])
		if err == nil {
			// parsing as an unlock hash was succesfull, store the pair and continue to the next pair
			pair.Condition = types.NewCondition(types.NewUnlockHashCondition(uh))
			pairs = append(pairs, pair)
			continue
		}

		// try to parse it as a JSON-encoded unlock condition
		err = pair.Condition.UnmarshalJSON([]byte(args[i]))
		if err != nil {
			err = fmt.Errorf("condition has to be UnlockHash or JSON-encoded UnlockCondition, output #%d's was neither", i/2)
			return
		}
		pairs = append(pairs, pair)
	}
	return
}
