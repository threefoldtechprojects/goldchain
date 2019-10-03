package explorer

import (
	"fmt"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"

	bolt "github.com/rivine/bbolt"
)

var (
	bucketInternal = []byte("Internal")
	// keys for bucketInternal
	internalBlockHeight  = []byte("BlockHeight")
	internalRecentChange = []byte("RecentChange")

	bucketChainFacts = []byte("ChainFacts")

	keyFactTotalSpentCoins  = []byte("SpentCoins")
	keyFactTotalLiquidCoins = []byte("LiquidCoins")
	keyFactTotalLockedCoins = []byte("Lockedcoins")
	keyFactTotalFeesPaid    = []byte("FeesPaid")
	keyFactTotalFeeDebt     = []byte("FeeDebt")
)

// dbSetInternal sets the specified key of bucketInternal to the encoded value.
func dbSetInternal(key []byte, val interface{}) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		valBytes, err := rivbin.Marshal(val)
		if err != nil {
			return fmt.Errorf("failed to (rivbin) marshal value: %v", err)
		}
		return tx.Bucket(bucketInternal).Put(key, valBytes)
	}
}

// dbGetInternal decodes the specified key of bucketInternal into the supplied pointer.
func dbGetInternal(key []byte, val interface{}) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		return rivbin.Unmarshal(tx.Bucket(bucketInternal).Get(key), val)
	}
}
