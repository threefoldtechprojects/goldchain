package modules

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type (
	// Wallet is an extended version of the regular Rivine Wallet
	Wallet interface {
		modules.Wallet

		// ConfirmedCustodyFeesToBePaid returns the total amount of custody fees to be paid
		// for all unlocked and locked confirmed coin outputs, in case you would spent them all.
		ConfirmedCustodyFeesToBePaid() (custodyfees types.Currency, err error)
	}
)