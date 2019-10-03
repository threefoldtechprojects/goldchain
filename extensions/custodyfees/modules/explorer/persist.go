package explorer

import (
	"os"
	"path/filepath"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

var explorerMetadata = persist.Metadata{
	Header:  "Custody Fee Explorer",
	Version: "1.0.0",
}

// initPersist initializes the persistent structures of the explorer module.
func (e *Explorer) initPersist(verbose bool) error {
	// Make the persist directory
	err := os.MkdirAll(e.persistDir, 0700)
	if err != nil {
		return err
	}

	// Initialize the logger.
	logFilePath := filepath.Join(e.persistDir, "explorer.log")
	e.log, err = persist.NewFileLogger(e.bcInfo, logFilePath, verbose)
	if err != nil {
		return err
	}

	// Open the database
	dbFilPath := filepath.Join(e.persistDir, "explorer.db")
	db, err := persist.OpenDatabase(explorerMetadata, dbFilPath)
	if err != nil {
		return err
	}
	e.db = db

	// Initialize the database
	err = e.db.Update(func(tx *bolt.Tx) error {
		internalBucket, err := tx.CreateBucketIfNotExists(bucketInternal)
		if err != nil {
			return err
		}

		// set default values for the bucketInternal
		blockHeightBytes, err := rivbin.Marshal(types.BlockHeight(0))
		if err != nil {
			return err
		}
		consensusChangeIDBytes, err := rivbin.Marshal(modules.ConsensusChangeID{})
		if err != nil {
			return err
		}
		internalDefaults := []struct {
			key, val []byte
		}{
			{internalBlockHeight, blockHeightBytes},
			{internalRecentChange, consensusChangeIDBytes},
		}
		for _, d := range internalDefaults {
			if internalBucket.Get(d.key) != nil {
				continue
			}
			err = internalBucket.Put(d.key, d.val)
			if err != nil {
				return err
			}
		}

		factsBucket, err := tx.CreateBucketIfNotExists(bucketChainFacts)
		if err != nil {
			return err
		}

		// set default values for the bucketChainFacts
		zeroCurrencyBytes, err := rivbin.Marshal(types.NewCurrency64(0))
		if err != nil {
			return err
		}
		factsDefaults := []struct {
			key, val []byte
		}{
			{keyFactTotalSpentCoins, zeroCurrencyBytes},
			{keyFactTotalLiquidCoins, zeroCurrencyBytes},
			{keyFactTotalLockedCoins, zeroCurrencyBytes},
			{keyFactTotalFeesPaid, zeroCurrencyBytes},
			{keyFactTotalFeeDebt, zeroCurrencyBytes},
		}
		for _, d := range factsDefaults {
			if factsBucket.Get(d.key) != nil {
				continue
			}
			err = factsBucket.Put(d.key, d.val)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
