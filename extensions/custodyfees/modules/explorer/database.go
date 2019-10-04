package explorer

import (
	"fmt"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

var (
	bucketInternal = []byte("Internal")
	// keys for bucketInternal
	internalBlockHeight  = []byte("BlockHeight")
	internalRecentChange = []byte("RecentChange")

	bucketMetrics = []byte("Metrics")

	keyMetricChainFacts = []byte("ChainFacts")

	bucketUnspentCoinOutputs = []byte("UnspentCoinOutputs")
	bucketSpentCoinOutputs   = []byte("SpentCoinOutputs")
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

func dbSetUnspentCoinOutput(bucket *bolt.Bucket, coid types.CoinOutputID, co types.CoinOutput) error {
	var lockValue uint64
	if co.Condition.ConditionType() == types.ConditionTypeTimeLock {
		lockValue = co.Condition.Condition.(*types.TimeLockCondition).LockTime
	}
	return dbSetUnspentCoinOutputWithLockTime(bucket, coid, lockValue)
}
func dbSetUnspentCoinOutputWithLockTime(bucket *bolt.Bucket, coid types.CoinOutputID, lockValue uint64) error {
	bID, err := rivbin.Marshal(coid)
	if err != nil {
		return err
	}
	bLockValue, err := rivbin.Marshal(lockValue)
	if err != nil {
		return err
	}
	return bucket.Put(bID, bLockValue)
}

func dbDeleteUnspentCoinOutput(bucket *bolt.Bucket, coid types.CoinOutputID) error {
	bID, err := rivbin.Marshal(coid)
	if err != nil {
		return err
	}
	return bucket.Delete(bID)
}

func dbMarkCoinOutputSpent(ucoBucket, scoBucket *bolt.Bucket, coid types.CoinOutputID) error {
	bID, err := rivbin.Marshal(coid)
	if err != nil {
		return err
	}
	b := ucoBucket.Get(bID)
	if len(b) == 0 {
		return fmt.Errorf("dbMarkCoinOutputSpent: failed to find lock value for unspent coin output %s", coid.String())
	}
	err = scoBucket.Put(bID, b)
	if err != nil {
		return err
	}
	return ucoBucket.Delete(bID)
}
func dbMarkCoinOutputUnspent(ucoBucket, scoBucket *bolt.Bucket, coid types.CoinOutputID) error {
	bID, err := rivbin.Marshal(coid)
	if err != nil {
		return err
	}
	b := scoBucket.Get(bID)
	if len(b) == 0 {
		return fmt.Errorf("dbMarkCoinOutputUnspent: failed to find lock value for spent coin output %s", coid.String())
	}
	err = ucoBucket.Put(bID, b)
	if err != nil {
		return err
	}
	return scoBucket.Delete(bID)
}

func dbGetUnspentCoinOutputLockValue(bucket *bolt.Bucket, coid types.CoinOutputID) (uint64, error) {
	bID, err := rivbin.Marshal(coid)
	if err != nil {
		return 0, err
	}
	b := bucket.Get(bID)
	if len(b) == 0 {
		return 0, fmt.Errorf("failed to find lock value for coin output %s", coid.String())
	}
	var lv uint64
	err = rivbin.Unmarshal(b, &lv)
	return lv, err
}

func dbUnspentCoinOutputValidatorMap(bucket *bolt.Bucket, f func(coid types.CoinOutputID, lockValue uint64) error) error {
	var (
		err       error
		coid      types.CoinOutputID
		lockValue uint64
	)
	return bucket.ForEach(func(bID, bLockValue []byte) error {
		err = rivbin.Unmarshal(bID, &coid)
		if err != nil {
			return err
		}
		err = rivbin.Unmarshal(bLockValue, &lockValue)
		if err != nil {
			return err
		}
		return f(coid, lockValue)
	})
}

func dbSetChainFactsDataFunc(facts ChainFacts) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		return dbSetChainFactsData(tx.Bucket(bucketMetrics), facts)
	}
}
func dbSetChainFactsData(bucket *bolt.Bucket, facts ChainFacts) error {
	b, err := rivbin.Marshal(facts)
	if err != nil {
		return fmt.Errorf("failed to (rivbin) marshal chain facts: %v", err)
	}
	return bucket.Put(keyMetricChainFacts, b)
}

func dbGetChainFactsDataFunc(facts *ChainFacts) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		return dbGetChainFactsData(tx.Bucket(bucketMetrics), facts)
	}
}
func dbGetChainFactsData(bucket *bolt.Bucket, facts *ChainFacts) error {
	return rivbin.Unmarshal(bucket.Get(keyMetricChainFacts), facts)
}

// ChainFacts collects all chain facts as one structure.
type ChainFacts struct {
	Height types.BlockHeight
	Time   types.Timestamp

	SpendableTokens       types.Currency
	SpendableLockedTokens types.Currency
	TotalCustodyFeeDebt   types.Currency

	SpentTokens     types.Currency
	PaidCustodyFees types.Currency
}
