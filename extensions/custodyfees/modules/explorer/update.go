package explorer

import (
	"fmt"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
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

		// Update cumulative stats for reverted blocks.
		for range cc.RevertedBlocks {
			blockheight--
		}

		// Update cumulative stats for applied blocks.
		for range cc.AppliedBlocks {
			blockheight++
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

		return nil
	})
	if err != nil {
		build.Critical("explorer update failed:", err)
	}
}

// helper functions
func assertNil(err error) {
	if err != nil {
		build.Critical(err)
	}
}
func assertRivMarshal(val interface{}) []byte {
	b, err := rivbin.Marshal(val)
	assertNil(err)
	return b
}
func mustPut(bucket *bolt.Bucket, key, val interface{}) {
	assertNil(bucket.Put(assertRivMarshal(key), assertRivMarshal(val)))
}
func mustPutSet(bucket *bolt.Bucket, key interface{}) {
	assertNil(bucket.Put(assertRivMarshal(key), nil))
}
func mustDelete(bucket *bolt.Bucket, key interface{}) {
	assertNil(bucket.Delete(assertRivMarshal(key)))
}
func bucketIsEmpty(bucket *bolt.Bucket) bool {
	k, _ := bucket.Cursor().First()
	return k == nil
}

// These functions panic on error. The panic will be caught by
// ProcessConsensusChange.

// TODO
