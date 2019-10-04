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

	allBuckets = [][]byte{
		bucketCoinOutputs,
	}
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

type (
	coinOutputDBInfo struct {
		CreationTime       types.Timestamp
		CreationValue      types.Currency
		FeeComputationTime types.Timestamp
		IsCustodyFee       bool
	}

	// CoinOutputInfo is all coin output info that can be requested from the plugin.
	CoinOutputInfo struct {
		CreationTime       types.Timestamp
		CreationValue      types.Currency
		IsCustodyFee       bool
		Spent              bool
		FeeComputationTime types.Timestamp
		CustodyFee         types.Currency
		SpendableValue     types.Currency
	}

	// CoinOutputInfoPreComputation is all coin output info that can be requested from the plugin,
	// minus the custody fee computation.
	CoinOutputInfoPreComputation struct {
		CreationTime       types.Timestamp
		CreationValue      types.Currency
		IsCustodyFee       bool
		Spent              bool
		FeeComputationTime types.Timestamp
	}
)

type (
	// CoinOutputInfoView allows you to view multiple coin outputs in a single View *bolt.Tx,
	// useful in case you want to get the info for multiple coin outputs.
	CoinOutputInfoView interface {
		// GetCoinOutputInfo returns the custody fee related coin output information for a given coin output ID,
		// returns an error only if the coin out never existed (spent or not).
		GetCoinOutputInfo(id types.CoinOutputID, chainTime types.Timestamp) (CoinOutputInfo, error)
		// GetCoinOutputInfoPreComputation returns the custody fee related coin output information for a given coin output ID,
		// returns an error only if the coin out never existed (spent or not).
		// Similar to `GetCoinOutputInfo` with the difference that the fee and spendable value aren't calculated yet.
		GetCoinOutputInfoPreComputation(id types.CoinOutputID) (CoinOutputInfoPreComputation, error)
	}

	txCoinOutputInfoView struct {
		rootBucket *bolt.Bucket
	}
)

func (view *txCoinOutputInfoView) GetCoinOutputInfo(id types.CoinOutputID, chainTime types.Timestamp) (CoinOutputInfo, error) {
	coBucket := view.rootBucket.Bucket(bucketCoinOutputs)
	if coBucket == nil {
		return CoinOutputInfo{}, fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs")
	}
	return getCoinOutputInfo(coBucket, id, chainTime)
}

func (view *txCoinOutputInfoView) GetCoinOutputInfoPreComputation(id types.CoinOutputID) (CoinOutputInfoPreComputation, error) {
	coBucket := view.rootBucket.Bucket(bucketCoinOutputs)
	if coBucket == nil {
		return CoinOutputInfoPreComputation{}, fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs")
	}
	return getCoinOutputInfoPreComputation(coBucket, id)
}

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

// GetCoinOutputInfo returns the custody fee related coin output information for a given coin output ID,
// returns an error only if the coin out never existed (spent or not).
func (p *Plugin) GetCoinOutputInfo(id types.CoinOutputID, chainTime types.Timestamp) (CoinOutputInfo, error) {
	var info CoinOutputInfo
	err := p.ViewCoinOutputInfo(func(view CoinOutputInfoView) error {
		var err error
		info, err = view.GetCoinOutputInfo(id, chainTime)
		return err
	})
	return info, err
}

// GetCoinOutputInfoPreComputation returns the custody fee related coin output information for a given coin output ID,
// returns an error only if the coin out never existed (spent or not).
// Similar to `GetCoinOutputInfo` with the difference that the fee and spendable value aren't calculated yet.
func (p *Plugin) GetCoinOutputInfoPreComputation(id types.CoinOutputID) (CoinOutputInfoPreComputation, error) {
	var info CoinOutputInfoPreComputation
	err := p.ViewCoinOutputInfo(func(view CoinOutputInfoView) error {
		var err error
		info, err = view.GetCoinOutputInfoPreComputation(id)
		return err
	})
	return info, err
}

// ViewCoinOutputInfo allows you to view the info for one or multiple coin outputs,
// in a single *bolt.Tx view.
func (p *Plugin) ViewCoinOutputInfo(f func(CoinOutputInfoView) error) error {
	return p.storage.View(func(rootBucket *bolt.Bucket) error {
		return f(&txCoinOutputInfoView{rootBucket})
	})
}

// InitPlugin initializes the Bucket for the first time
func (p *Plugin) InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, storage modules.PluginViewStorage, unregisterCallback modules.PluginUnregisterCallback) (persist.Metadata, error) {
	p.storage = storage
	p.unregisterCallback = unregisterCallback
	if metadata == nil {
		for _, bucketName := range allBuckets {
			subBucket := bucket.Bucket([]byte(bucketName))
			if subBucket == nil {
				var err error
				_, err = bucket.CreateBucket([]byte(bucketName))
				if err != nil {
					return persist.Metadata{}, fmt.Errorf("failed to create %s bucket for custody fees plugin: %v", string(bucketName), err)
				}
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
	for idx, mp := range block.MinerPayouts {
		mpid := types.CoinOutputID(block.MinerPayoutID(uint64(idx)))
		bMPID, err := rivbin.Marshal(mpid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal (miner payout ID as) coin output ID: %v", err)
		}
		bInfo, err := rivbin.Marshal(coinOutputDBInfo{
			CreationTime:       block.Timestamp,
			CreationValue:      mp.Value,
			FeeComputationTime: 0,
			IsCustodyFee:       false,
		})
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal miner payout info: %v", err)
		}
		err = coBucket.Put(bMPID, bInfo)
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
		err = p.applyTransaction(cTxn, coBucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyBlockHeader applies data from a block header to the custodyfee bucket.
func (p *Plugin) ApplyBlockHeader(header modules.ConsensusBlockHeader, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("custodyfee bucket does not exist")
	}
	// apply miner payouts
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}
	for idx, mpid := range header.MinerPayoutIDs {
		bMPID, err := rivbin.Marshal(mpid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal (miner payout ID as) coin output ID: %v", err)
		}
		bInfo, err := rivbin.Marshal(coinOutputDBInfo{
			CreationTime:       header.Timestamp,
			CreationValue:      header.MinerPayouts[idx].Value,
			FeeComputationTime: 0,
			IsCustodyFee:       false,
		})
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal miner payout info: %v", err)
		}
		err = coBucket.Put(bMPID, bInfo)
		if err != nil {
			return fmt.Errorf("failed to link (miner payout ID as) coin output's ID to its block time: %v", err)
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
	return p.applyTransaction(txn, coBucket)
}

func (p *Plugin) applyTransaction(txn modules.ConsensusTransaction, coBucket *bolt.Bucket) error {
	var computationTime types.Timestamp
	for index, co := range txn.CoinOutputs {
		isCustodyFee := co.Condition.ConditionType() == cftypes.ConditionTypeCustodyFee
		if isCustodyFee {
			computationTime = co.Condition.Condition.(*cftypes.CustodyFeeCondition).ComputationTime
		}
		coid := txn.CoinOutputID(uint64(index))
		bCOID, err := rivbin.Marshal(coid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output ID: %v", err)
		}
		bInfo, err := rivbin.Marshal(coinOutputDBInfo{
			CreationTime:       txn.BlockTime,
			CreationValue:      co.Value,
			FeeComputationTime: 0,
			IsCustodyFee:       isCustodyFee,
		})
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output info: %v", err)
		}
		err = coBucket.Put(bCOID, bInfo)
		if err != nil {
			return fmt.Errorf("failed to link coin output's ID to its block time: %v", err)
		}
	}
	for _, ci := range txn.CoinInputs {
		currentInfo, err := getCoinOutputInfoPreComputation(coBucket, ci.ParentID)
		if err != nil {
			return fmt.Errorf("failed to look up coin input %s in custody fees DB: %v", ci.ParentID.String(), err)
		}
		bCOID, err := rivbin.Marshal(ci.ParentID)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output (used as coin input) ID: %v", err)
		}
		bInfo, err := rivbin.Marshal(coinOutputDBInfo{
			CreationTime:       currentInfo.CreationTime,
			CreationValue:      currentInfo.CreationValue,
			FeeComputationTime: computationTime,
			IsCustodyFee:       false,
		})
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output (used as coin input) info: %v", err)
		}
		err = coBucket.Put(bCOID, bInfo)
		if err != nil {
			return fmt.Errorf("failed to link coin input's ID to its block time: %v", err)
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
		mpid := block.MinerPayoutID(uint64(idx))
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

// RevertBlockHeader reverts data from a block header from the custodyfee bucket.
func (p *Plugin) RevertBlockHeader(header modules.ConsensusBlockHeader, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("custodyfee bucket does not exist")
	}
	if len(header.MinerPayouts) == 0 {
		return nil // nothing to do
	}
	coBucket, err := bucket.Bucket(bucketCoinOutputs)
	if err != nil {
		return fmt.Errorf("corrupt custody fee plugin: did not find any coin outputs: %v", err)
	}
	for _, mpid := range header.MinerPayoutIDs {
		bMPID, err := rivbin.Marshal(mpid)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal (miner payout ID as) coin output ID: %v", err)
		}
		err = coBucket.Delete(bMPID)
		if err != nil {
			return fmt.Errorf("failed to unlink (miner payout ID as) coin output's ID from its block time: %v", err)
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
	for _, ci := range txn.CoinInputs {
		currentInfo, err := getCoinOutputInfoPreComputation(coBucket, ci.ParentID)
		if err != nil {
			return fmt.Errorf("failed to look up coin input %s in custody fees DB: %v", ci.ParentID.String(), err)
		}
		bCOID, err := rivbin.Marshal(ci.ParentID)
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output (used as coin input) ID: %v", err)
		}
		bInfo, err := rivbin.Marshal(coinOutputDBInfo{
			CreationTime:       currentInfo.CreationTime,
			CreationValue:      currentInfo.CreationValue,
			FeeComputationTime: 0,
			IsCustodyFee:       false,
		})
		if err != nil {
			return fmt.Errorf("failed to rivbin marshal coin output (used as coin input) info: %v", err)
		}
		err = coBucket.Put(bCOID, bInfo)
		if err != nil {
			return fmt.Errorf("failed to link coin input's ID to its block time: %v", err)
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
		info, err := getCoinOutputInfo(coBucket, ci.ParentID, ctx.BlockTime)
		if err != nil {
			return err
		}
		if info.Spent {
			return fmt.Errorf("coin output %s is already marked as spent in the custody fees DB: cannot be spend again", ci.ParentID.String())
		}
		requiredCustodyFee = requiredCustodyFee.Add(info.CustodyFee)
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

func getCoinOutputInfoPreComputation(coBucket *bolt.Bucket, id types.CoinOutputID) (CoinOutputInfoPreComputation, error) {
	bID, err := rivbin.Marshal(id)
	if err != nil {
		return CoinOutputInfoPreComputation{}, fmt.Errorf("failed to rivbin marshal coin input parent ID: %v", err)
	}

	b := coBucket.Get(bID)
	if len(b) == 0 {
		return CoinOutputInfoPreComputation{}, fmt.Errorf("failed to find timestamp in CustodyFee DB for coin input %s", id.String())
	}

	var dbInfo coinOutputDBInfo
	err = rivbin.Unmarshal(b, &dbInfo)
	if err != nil {
		return CoinOutputInfoPreComputation{}, fmt.Errorf(
			"failed to unmarshal coin output %s's info: %v",
			id.String(), err)
	}

	return CoinOutputInfoPreComputation{
		CreationTime:       dbInfo.CreationTime,
		CreationValue:      dbInfo.CreationValue,
		IsCustodyFee:       dbInfo.IsCustodyFee,
		Spent:              !dbInfo.IsCustodyFee && dbInfo.FeeComputationTime > 0,
		FeeComputationTime: dbInfo.FeeComputationTime,
	}, nil
}

func getCoinOutputInfo(coBucket *bolt.Bucket, id types.CoinOutputID, chainTime types.Timestamp) (CoinOutputInfo, error) {
	var info CoinOutputInfo
	preComputationInfo, err := getCoinOutputInfoPreComputation(coBucket, id)
	if err != nil {
		return info, err
	}
	info.CreationTime = preComputationInfo.CreationTime
	info.CreationValue = preComputationInfo.CreationValue
	info.IsCustodyFee = preComputationInfo.IsCustodyFee
	if info.IsCustodyFee {
		return info, err // no fee is required, and nothing of it is spendable
	}
	if preComputationInfo.FeeComputationTime == 0 {
		if info.CreationTime > chainTime {
			return info, fmt.Errorf(
				"unspent coin output %s is created in the future (%d) compared to given chain time %d",
				id.String(), info.CreationTime, chainTime)
		}
		info.FeeComputationTime = chainTime
	} else {
		info.Spent = true
		info.FeeComputationTime = preComputationInfo.FeeComputationTime
	}
	if info.FeeComputationTime != info.CreationTime {
		info.SpendableValue, info.CustodyFee = AmountCustodyFeePairAfterXSeconds(info.CreationValue, info.CreationTime-info.FeeComputationTime)
	} else {
		info.SpendableValue = info.CreationValue
	}
	return info, nil
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	return p.storage.Close()
}
