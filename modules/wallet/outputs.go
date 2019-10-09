package wallet

import (
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	gcmodules "github.com/nbh-digital/goldchain/modules"
)

// UnlockedUnspendOutputs returns all unlocked coinoutput and blockstakeoutputs
func (w *Wallet) UnlockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	ucom := make(map[types.CoinOutputID]types.CoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all coin and block stake stum
	for id, co := range w.coinOutputs {
		if co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// same for multisig
	for id, co := range w.multiSigCoinOutputs {
		if co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// block stakes
	for id, bso := range w.blockstakeOutputs {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	// block stake multisigs
	for id, bso := range w.multiSigBlockStakeOutputs {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	return ucom, ubsom, nil
}

// LockedUnspendOutputs returnas all locked coinoutput and blockstakeoutputs
func (w *Wallet) LockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	ucom := make(map[types.CoinOutputID]types.CoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all coin and block stake stum
	for id, co := range w.coinOutputs {
		if !co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// same for multisig
	for id, co := range w.multiSigCoinOutputs {
		if !co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// block stakes
	for id, bso := range w.blockstakeOutputs {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	// block stake multisigs
	for id, bso := range w.multiSigBlockStakeOutputs {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	return ucom, ubsom, nil
}

// UnlockedUnspendOutputsWithCustodyFeeInfo returns all unlocked and unspend coin and blockstake outputs
// owned by this wallet, adding custody fee information to each spent unlocked coin output.
func (w *Wallet) UnlockedUnspendOutputsWithCustodyFeeInfo() (map[types.CoinOutputID]gcmodules.WalletCoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	ucom := make(map[types.CoinOutputID]gcmodules.WalletCoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	var (
		err  error
		info custodyfees.CoinOutputInfo
	)
	err = w.cfplugin.ViewCoinOutputInfo(func(view custodyfees.CoinOutputInfoView) error {
		// get all coin and block stake stum
		for id, co := range w.coinOutputs {
			if co.Condition.Fulfillable(ctx) {
				info, err = view.GetCoinOutputInfo(id, ctx.BlockTime)
				if err != nil {
					return fmt.Errorf("failed to get custodyfee info for unspent unlocked coin output %s: %v", id.String(), err)
				}
				ucom[id] = gcmodules.WalletCoinOutput{
					CoinOutput: co,
					CoinInfo:   info,
				}
			}
		}
		// same for multisig
		for id, co := range w.multiSigCoinOutputs {
			if co.Condition.Fulfillable(ctx) {
				info, err = view.GetCoinOutputInfo(id, ctx.BlockTime)
				if err != nil {
					return fmt.Errorf("failed to get custodyfee info for unspent unlocked ms coin output %s: %v", id.String(), err)
				}
				ucom[id] = gcmodules.WalletCoinOutput{
					CoinOutput: co,
					CoinInfo:   info,
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// block stakes
	for id, bso := range w.blockstakeOutputs {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	// block stake multisigs
	for id, bso := range w.multiSigBlockStakeOutputs {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	return ucom, ubsom, nil
}

// LockedUnspendOutputsWithCustodyFeeInfo returns all locked and unspend coin and blockstake outputs owned
// by this wallet, adding custody fee information to each spent locked coin output.
func (w *Wallet) LockedUnspendOutputsWithCustodyFeeInfo() (map[types.CoinOutputID]gcmodules.WalletCoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	ucom := make(map[types.CoinOutputID]gcmodules.WalletCoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	var (
		err  error
		info custodyfees.CoinOutputInfo
	)
	err = w.cfplugin.ViewCoinOutputInfo(func(view custodyfees.CoinOutputInfoView) error {
		// get all coin and block stake stum
		for id, co := range w.coinOutputs {
			if !co.Condition.Fulfillable(ctx) {
				info, err = view.GetCoinOutputInfo(id, ctx.BlockTime)
				if err != nil {
					return fmt.Errorf("failed to get custodyfee info for unspent locked coin output %s: %v", id.String(), err)
				}
				ucom[id] = gcmodules.WalletCoinOutput{
					CoinOutput: co,
					CoinInfo:   info,
				}
			}
		}
		// same for multisig
		for id, co := range w.multiSigCoinOutputs {
			if !co.Condition.Fulfillable(ctx) {
				info, err = view.GetCoinOutputInfo(id, ctx.BlockTime)
				if err != nil {
					return fmt.Errorf("failed to get custodyfee info for unspent locked ms coin output %s: %v", id.String(), err)
				}
				ucom[id] = gcmodules.WalletCoinOutput{
					CoinOutput: co,
					CoinInfo:   info,
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// block stakes
	for id, bso := range w.blockstakeOutputs {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	// block stake multisigs
	for id, bso := range w.multiSigBlockStakeOutputs {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	return ucom, ubsom, nil
}
