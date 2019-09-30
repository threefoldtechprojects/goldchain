package custodyfees

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	cftypes "github.com/nbh-digital/goldchain/extensions/custodyfees/types"

	bolt "github.com/rivine/bbolt"
)

const (
	pluginDBVersion = "1.0.0.0"
	pluginDBHeader  = "custodyFeePlugin"
)

var (
	bucketCoinOutputs = []byte("coinoutputs")
)

type (
	// Plugin is a struct defines the custodyfee plugin
	Plugin struct {
		storage            modules.PluginViewStorage
		unregisterCallback modules.PluginUnregisterCallback

		binMarshal   func(v interface{}) ([]byte, error)
		binUnmarshal func(b []byte, v interface{}) error
	}
)

// TODO:
// - validate all transactions to ensure that transactions that contain coin inputs, also define a coin output with the custody fee condition

// NewPlugin creates a new CustodyFee Plugin,
// also registering the condition type
func NewPlugin() *Plugin {
	types.RegisterUnlockConditionType(cftypes.ConditionTypeCustodyFee, func() types.MarshalableUnlockCondition { return &cftypes.CustodyFeeCondition{} })
	return new(Plugin)
}

// InitPlugin initializes the Bucket for the first time
func (p *Plugin) InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, storage modules.PluginViewStorage, unregisterCallback modules.PluginUnregisterCallback) (persist.Metadata, error) {
	p.storage = storage
	p.unregisterCallback = unregisterCallback
	if metadata == nil {
		coinOutputsBucket := bucket.Bucket([]byte(bucketCoinOutputs))
		if coinOutputsBucket == nil {
			var err error
			coinOutputsBucket, err = bucket.CreateBucket([]byte(bucketCoinOutputs))
			if err != nil {
				return persist.Metadata{}, fmt.Errorf("failed to create coin outputs bucket: %v", err)
			}
		}

		metadata = &persist.Metadata{
			Version: pluginDBVersion,
			Header:  pluginDBHeader,
		}
	} else if metadata.Version != pluginDBVersion {
		return persist.Metadata{}, errors.New("There is only 1 version of this plugin, version mismatch")
	}
	return *metadata, nil
}

// ApplyBlock applies a block's custodyfee transactions to the custodyfee bucket.
func (p *Plugin) ApplyBlock(block modules.ConsensusBlock, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("custodyfee bucket does not exist")
	}
	var err error
	for idx, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			BlockHeight:            block.Height,
			BlockTime:              block.Timestamp,
			SequenceID:             uint16(idx),
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.ApplyTransaction(cTxn, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyTransaction applies a custodyfee transactions to the custodyfee bucket.
func (p *Plugin) ApplyTransaction(txn modules.ConsensusTransaction, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("custodyfee bucket does not exist")
	}
	if len(txn.CoinOutputs) == 0 {
		return nil // nothing to do
	}
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs")
	}
	btValue, err := rivbin.Marshal(txn.BlockTime)
	if err != nil {
		return fmt.Errorf("failed to rivbin marshal block time: %v", err)
	}
	for index := range txn.CoinOutputs {
		coid := txn.CoinOutputID(uint64(index))
		bCOID, err := rivbin.Marshal(coid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output ID: %v", err)
		}
		err = coBucket.Put(bCOID, btValue)
		if err != nil {
			return fmt.Errorf("failed to link coin output's ID to its block time: %v", err)
		}
	}
	return nil
}

// RevertBlock reverts a block's custodyfee transaction from the custodyfee bucket
func (p *Plugin) RevertBlock(block modules.ConsensusBlock, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("mint conditions bucket does not exist")
	}
	// collect all one-per-block mint conditions
	var err error
	for idx, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			BlockHeight:            block.Height,
			BlockTime:              block.Timestamp,
			SequenceID:             uint16(idx),
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.RevertTransaction(cTxn, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevertTransaction reverts a custodyfee transactions to the custodyfee bucket.
func (p *Plugin) RevertTransaction(txn modules.ConsensusTransaction, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("custodyfee bucket does not exist")
	}
	if len(txn.CoinOutputs) == 0 {
		return nil // nothing to do
	}
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs")
	}
	for index := range txn.CoinOutputs {
		coid := txn.CoinOutputID(uint64(index))
		bCOID, err := rivbin.Marshal(coid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output ID: %v", err)
		}
		err = coBucket.Delete(bCOID)
		if err != nil {
			return fmt.Errorf("failed to unlink coin output's ID from its block time: %v", err)
		}
	}
	return nil
}

// TransactionValidatorVersionFunctionMapping returns all tx validators linked to this plugin
func (p *Plugin) TransactionValidatorVersionFunctionMapping() map[types.TransactionVersion][]modules.PluginTransactionValidationFunction {
	return nil
}

// TransactionValidators returns all tx validators linked to this plugin
func (p *Plugin) TransactionValidators() []modules.PluginTransactionValidationFunction {
	return nil
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	return p.storage.Close()
}
