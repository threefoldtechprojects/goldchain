package client

import (
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/types"
	"github.com/threefoldtech/rivine/pkg/client"
)

func parseCoinArg(cc client.CurrencyConvertor, str string) types.Currency {
	amount, err := cc.ParseCoinString(str)
	if err != nil {
		fmt.Fprintln(os.Stderr, cc.CoinArgDescription("amount"))
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse coin-typed argument: ", err)
		return types.Currency{}
	}
	return amount
}
