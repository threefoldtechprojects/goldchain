package explorer

import (
	"fmt"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
)

// ProcessConsensusChange follows the most recent changes to the consensus set,
// including parsing new blocks and updating the utxo sets.
func (e *Explorer) ProcessConsensusChange(cc modules.ConsensusChange) {
	if len(cc.AppliedBlocks) == 0 {
		build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
	}

	err := e.db.Update(func(tx *bolt.Tx) (err error) {
		// use exception-style error handling to enable more concise update code
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()

		// get starting block height
		var blockheight types.BlockHeight
		err = dbGetInternal(internalBlockHeight, &blockheight)(tx)
		if err != nil {
			return err
		}

		// get unspentCoinOutputsBucket and bucketSpentCoinOutputs to update it
		ucoBucket := tx.Bucket(bucketUnspentCoinOutputs)
		if ucoBucket != nil {
			return fmt.Errorf("corrupt Custody Fee Explorer: did not find bucket %s", string(bucketUnspentCoinOutputs))
		}
		scoBucket := tx.Bucket(bucketSpentCoinOutputs)
		if ucoBucket != nil {
			return fmt.Errorf("corrupt Custody Fee Explorer: did not find bucket %s", string(bucketSpentCoinOutputs))
		}

		var coid types.CoinOutputID
		revertedCoinInputIDs := map[types.CoinOutputID]types.Timestamp{}
		var blocktime types.Timestamp

		// Update cumulative stats for reverted blocks.
		for _, block := range cc.RevertedBlocks {
			blockheight--
			blocktime = block.Timestamp
			for idx := range block.MinerPayouts {
				coid = block.MinerPayoutID(uint64(idx))
				err = dbDeleteUnspentCoinOutput(ucoBucket, coid)
				if err != nil {
					return err
				}
				revertedCoinInputIDs[coid] = blocktime
			}
			for _, txn := range block.Transactions {
				for _, ci := range txn.CoinInputs {
					err = dbMarkCoinOutputUnspent(ucoBucket, scoBucket, ci.ParentID)
					if err != nil {
						return err
					}
				}
				for idx := range txn.CoinOutputs {
					coid = txn.CoinOutputID(uint64(idx))
					err = dbDeleteUnspentCoinOutput(ucoBucket, coid)
					revertedCoinInputIDs[coid] = blocktime
				}
			}
		}

		// Update cumulative stats for applied blocks.
		for _, block := range cc.AppliedBlocks {
			blockheight++
			blocktime = block.Timestamp
			for idx := range block.MinerPayouts {
				coid = block.MinerPayoutID(uint64(idx))
				err = dbSetUnspentCoinOutputWithLockTime(ucoBucket, coid, uint64(blockheight+e.chainCts.MaturityDelay))
				if err != nil {
					return err
				}
			}
			for _, txn := range block.Transactions {
				for _, ci := range txn.CoinInputs {
					err = dbMarkCoinOutputSpent(ucoBucket, scoBucket, ci.ParentID)
					if err != nil {
						return err
					}
				}
				for idx, co := range txn.CoinOutputs {
					coid = txn.CoinOutputID(uint64(idx))
					err = dbSetUnspentCoinOutput(ucoBucket, coid, co)
				}
			}
		}

		// set final blockheight
		err = dbSetInternal(internalBlockHeight, blockheight)(tx)
		if err != nil {
			return err
		}

		// set change ID
		err = dbSetInternal(internalRecentChange, cc.ID)(tx)
		if err != nil {
			return err
		}

		// get bucketMetrics to update it
		metricsBucket := tx.Bucket(bucketMetrics)
		if metricsBucket != nil {
			return fmt.Errorf("corrupt Custody Fee Explorer: did not find bucket %s", string(bucketMetrics))
		}

		// get current chain stats
		var facts ChainFacts
		err = dbGetChainFactsData(metricsBucket, &facts)
		if err != nil {
			return err
		}

		// compute new chain facts
		// ... set height/time info
		facts.Height = blockheight
		facts.Time = blocktime
		// ... set aggregated currency info
		err = e.plugin.ViewCoinOutputInfo(func(view custodyfees.CoinOutputInfoView) error {
			var info custodyfees.CoinOutputInfo
			// first subtract all aggregated spent/paid values
			for coid, chainTime := range revertedCoinInputIDs {
				info, err = view.GetCoinOutputInfo(coid, chainTime)
				if err != nil {
					return err
				}
				if info.FeeComputationTime != chainTime {
					e.log.Printf("[WARN] unexpected fee computation time for reverted spent coin output %s: %d != %d", coid.String(), info.FeeComputationTime, chainTime)
				}
				facts.SpentTokens = facts.SpentTokens.Sub(info.SpendableValue)
				facts.PaidCustodyFees = facts.PaidCustodyFees.Sub(info.CustodyFee)
			}
			// recalculate the liquid, locked (both spendable) and fee debt
			facts.SpendableTokens = types.Currency{}
			facts.SpendableLockedTokens = types.Currency{}
			facts.TotalCustodyFeeDebt = types.Currency{}
			return dbUnspentCoinOutputValidatorMap(ucoBucket, func(coid types.CoinOutputID, lockValue uint64) error {
				// get locked state
				var locked bool
				if lockValue > 0 {
					if lockValue < types.LockTimeMinTimestampValue {
						locked = types.BlockHeight(lockValue) < blockheight
					} else {
						locked = types.Timestamp(lockValue) < blocktime
					}
				}
				// get spendable and custody fee
				info, err = view.GetCoinOutputInfo(coid, blocktime)
				if err != nil {
					return err
				}
				// log some stuff that are not critical but might slightly mess up the aggregated sats
				if info.Spent {
					e.log.Printf("[WARN] unexpected spent sate for coin output %s: will still be counted as unspent with wrong values for now", coid.String())
				}
				// update aggregated sats
				facts.TotalCustodyFeeDebt = facts.TotalCustodyFeeDebt.Add(info.CustodyFee)
				if locked {
					facts.SpendableLockedTokens = facts.SpendableLockedTokens.Add(info.SpendableValue)
				} else {
					facts.SpendableTokens = facts.SpendableTokens.Add(info.SpendableValue)
				}
				// all good
				return nil
			})
		})
		if err != nil {
			return err
		}

		// set update chain stats
		err = dbSetChainFactsData(metricsBucket, facts)
		if err != nil {
			return err
		}

		// all good
		return nil
	})
	if err != nil {
		build.Critical("explorer update failed:", err)
	}
}
