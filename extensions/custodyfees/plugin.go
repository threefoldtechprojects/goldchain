package custodyfees

import (
	"encoding/hex"
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
		maxAllowedComputationTimeAdvance types.Timestamp

		storage            modules.PluginViewStorage
		unregisterCallback modules.PluginUnregisterCallback

		binMarshal   func(v interface{}) ([]byte, error)
		binUnmarshal func(b []byte, v interface{}) error
	}
)

// TODO:
// - ensure plugin is aware of minerfee payouts at all times... (Rivine change that requires update in extensions as well as tfchain!!!)
//		- this can be done by introducing Apply/Revert BlockHeader...
//			(this is required as actual applied txs are done via ApplyTransaction, not via ApplyBlock, at least in the regular path...)
// - store more info: spentTime (== computationRegistration, if 0 -> unspent), value
// 		- also return creation time
// 		- improve explorer backend based on this change

// NewPlugin creates a new CustodyFee Plugin,
// also registering the condition type
func NewPlugin(maxAllowedComputationTimeAdvance types.Timestamp) *Plugin {
	if maxAllowedComputationTimeAdvance == 0 {
		panic("maxAllowedComputationTimeAdvance has to have a value greater than 0")
	}
	types.RegisterUnlockConditionType(cftypes.ConditionTypeCustodyFee, func() types.MarshalableUnlockCondition { return &cftypes.CustodyFeeCondition{} })
	return &Plugin{
		maxAllowedComputationTimeAdvance: maxAllowedComputationTimeAdvance,
	}
}

// GetCoinOutputCreationTime returns the timestamp of creation of a coin output
func (p *Plugin) GetCoinOutputCreationTime(id types.CoinOutputID) (types.Timestamp, error) {
	var ts types.Timestamp
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		coBucket := bucket.Bucket(bucketCoinOutputs)
		if coBucket == nil {
			return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs")
		}
		var err error
		ts, err = getCoinOutputTime(coBucket, id)
		if err != nil {
			return fmt.Errorf("failed to look up creation timing of coin output %s: %v", id.String(), err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return ts, nil
}

// InitPlugin initializes the Bucket for the first time
func (p *Plugin) InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, storage modules.PluginViewStorage, unregisterCallback modules.PluginUnregisterCallback) (persist.Metadata, error) {
	p.storage = storage
	p.unregisterCallback = unregisterCallback
	if metadata == nil {
		coinOutputsBucket := bucket.Bucket([]byte(bucketCoinOutputs))
		if coinOutputsBucket == nil {
			var err error
			_, err = bucket.CreateBucket([]byte(bucketCoinOutputs))
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
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}
	btValue, err := rivbin.Marshal(block.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to rivbin marshal block time: %v", err)
	}
	for idx := range block.MinerPayouts {
		mpid := types.CoinOutputID(block.MinerPayoutID(uint64(idx)))
		bMPID, err := rivbin.Marshal(mpid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal (miner payout ID as) coin output ID: %v", err)
		}
		err = coBucket.Put(bMPID, btValue)
		if err != nil {
			return fmt.Errorf("failed to link (miner payout ID as) coin output's ID to its block time: %v", err)
		}
	}
	for idx, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			BlockHeight:            block.Height,
			BlockTime:              block.Timestamp,
			SequenceID:             uint16(idx),
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.applyTransaction(cTxn, coBucket, btValue)
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
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}
	btValue, err := rivbin.Marshal(txn.BlockTime)
	if err != nil {
		return fmt.Errorf("failed to rivbin marshal block time: %v", err)
	}
	return p.applyTransaction(txn, coBucket, btValue)
}

func (p *Plugin) applyTransaction(txn modules.ConsensusTransaction, coBucket *bolt.Bucket, blockTime []byte) error {
	if len(txn.CoinOutputs) == 0 {
		return nil // nothing to do
	}
	for index := range txn.CoinOutputs {
		coid := txn.CoinOutputID(uint64(index))
		bCOID, err := rivbin.Marshal(coid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output ID: %v", err)
		}
		err = coBucket.Put(bCOID, blockTime)
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
	var err error
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}
	for idx := range block.MinerPayouts {
		mpid := types.CoinOutputID(block.MinerPayoutID(uint64(idx)))
		bMPID, err := rivbin.Marshal(mpid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal (miner payout ID as) coin output ID: %v", err)
		}
		err = coBucket.Delete(bMPID)
		if err != nil {
			return fmt.Errorf("failed to unlink (miner payout ID as) coin output's ID from its block time: %v", err)
		}
	}
	for idx, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			BlockHeight:            block.Height,
			BlockTime:              block.Timestamp,
			SequenceID:             uint16(idx),
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.revertTransaction(cTxn, coBucket)
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
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}
	return p.revertTransaction(txn, coBucket)
}

func (p *Plugin) revertTransaction(txn modules.ConsensusTransaction, coBucket *bolt.Bucket) error {
	if len(txn.CoinOutputs) == 0 {
		return nil // nothing to do
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
	return []modules.PluginTransactionValidationFunction{
		p.validateCustodyFeePresent,
	}
}

func (p *Plugin) validateCustodyFeePresent(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	if len(tx.CoinInputs) == 0 {
		return nil // nothing to do
	}

	// ensure there is one (and only one) custody fee condition,
	// that is within an accepted timeframe
	var (
		computationTime types.Timestamp
		custodyFeeValue types.Currency
	)
	for _, co := range tx.CoinOutputs {
		if co.Condition.ConditionType() != cftypes.ConditionTypeCustodyFee {
			continue
		}
		cfc, ok := co.Condition.Condition.(*cftypes.CustodyFeeCondition)
		if !ok {
			return fmt.Errorf("unexpected unlock condition for condition type %d", co.Condition.ConditionType())
		}
		if computationTime != 0 {
			return errors.New("only one custody fee condition per Tx is allowed")
		}
		computationTime = cfc.ComputationTime
		custodyFeeValue = co.Value
	}
	if computationTime == 0 {
		return errors.New("tx does not contain the required coin output for the custody fee, while coin inputs are spent")
	}
	if ctx.BlockTime < computationTime {
		return errors.New("registered custody fee computation time cannot be in the future")
	}
	if diff := ctx.BlockTime - computationTime; diff > p.maxAllowedComputationTimeAdvance {
		return fmt.Errorf(
			"custody fee is paid, computated based on a timestamp too far in the past: %ds too late",
			diff-p.maxAllowedComputationTimeAdvance)
	}

	// get coin out bucket,
	// where all known coin outputs are linked to the timing they are created
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}

	// computate required custody fee
	var requiredCustodyFee types.Currency
	// ... look up each coin input in our plugin DB,
	//     to check how much the fee will cost
	for _, ci := range tx.CoinInputs {
		ciTS, err := getCoinOutputTime(coBucket, ci.ParentID)
		if err != nil {
			return fmt.Errorf("failed to look up creation timing of coin input %s: %v", ci.ParentID.String(), err)
		}
		if ciTS > ctx.BlockTime {
			return fmt.Errorf("spent coin output creation time is in the future, this is invalid: %d > %d", ciTS, ctx.BlockTime)
		}
		if ciTS == ctx.BlockTime {
			continue // nothing to do
		}
		_, fee := AmountCustodyFeePairAfterXSeconds(tx.SpentCoinOutputs[ci.ParentID].Value, ctx.BlockTime-ciTS)
		requiredCustodyFee = requiredCustodyFee.Add(fee)
	}

	// ensure the custody fee is exactly as expected
	if !requiredCustodyFee.Equals(custodyFeeValue) {
		return fmt.Errorf(
			"unexpected custody fee of value %s expected %s",
			custodyFeeValue.String(), requiredCustodyFee.String())
	}

	// transaction is valid
	return nil
}

func getCoinOutputTime(coBucket *bolt.Bucket, id types.CoinOutputID) (types.Timestamp, error) {
	bID, err := rivbin.Marshal(id)
	if err != nil {
		return 0, fmt.Errorf("failed to rivbin marshal coin input parent ID: %v", err)
	}

	b := coBucket.Get(bID)
	if len(b) == 0 {
		return 0, fmt.Errorf("failed to find timestamp in CustodyFee DB for coin input %s", id.String())
	}

	var ts types.Timestamp
	err = rivbin.Unmarshal(b, &ts)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to unmarshal coin output %s's timestamp 0x%s: %v",
			id.String(), hex.EncodeToString(b), err)
	}

	return ts, nil
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	return p.storage.Close()
}
