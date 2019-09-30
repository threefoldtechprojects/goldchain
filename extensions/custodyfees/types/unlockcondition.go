package types

import (
	"github.com/threefoldtech/rivine/types"
)

const (
	// ConditionTypeCustodyFee defines the CustodyFeeCondition,
	// a condition used to defined the custody fees paid for every spent coin output.
	//
	// Implemented by the CustodyFeeCondition type.
	ConditionTypeCustodyFee types.ConditionType = 128
)

const (
	// UnlockTypeCustodyFee is the unlock type of the unlock hash used for the CustodyFee condition.
	UnlockTypeCustodyFee types.UnlockType = 128
)

// CustodyFeeUnlockHash is the address used for the custody fee condition.
var CustodyFeeUnlockHash = types.UnlockHash{Type: UnlockTypeCustodyFee}

// CustodyFeeCondition implements the ConditionTypeCustodyFee (unlock) ConditionType.
// See ConditionTypeCustodyFee for more information.
type CustodyFeeCondition struct {
	ComputationTime types.Timestamp `json:"computationtime"`
} // cannot be fulliled

// Fulfill implements UnlockCondition.Fulfill
func (cf *CustodyFeeCondition) Fulfill(fulfillment types.UnlockFulfillment, ctx types.FulfillContext) error {
	return types.ErrUnexpectedUnlockFulfillment // CustodyFeeCondition cannot be fulfilled

}

// ConditionType implements UnlockCondition.ConditionType
func (cf *CustodyFeeCondition) ConditionType() types.ConditionType { return ConditionTypeCustodyFee }

// IsStandardCondition implements UnlockCondition.IsStandardCondition
func (cf *CustodyFeeCondition) IsStandardCondition(types.ValidationContext) error { return nil } // always valid

// UnlockHash implements UnlockCondition.UnlockHash
func (cf *CustodyFeeCondition) UnlockHash() types.UnlockHash { return CustodyFeeUnlockHash }

// Equal implements UnlockCondition.Equal
func (cf *CustodyFeeCondition) Equal(c types.UnlockCondition) bool {
	if c == nil {
		return true // implicit equality
	}
	cfr, ok := c.(*CustodyFeeCondition)
	if !ok {
		return false
	}
	return cf.ComputationTime == cfr.ComputationTime
}

// Fulfillable implements UnlockCondition.Fulfillable
func (cf *CustodyFeeCondition) Fulfillable(types.FulfillableContext) bool { return false }

// Marshal implements MarshalableUnlockCondition.Marshal
func (cf *CustodyFeeCondition) Marshal(f types.MarshalFunc) ([]byte, error) {
	return f(cf.ComputationTime)
}

// Unmarshal implements MarshalableUnlockCondition.Unmarshal
func (cf *CustodyFeeCondition) Unmarshal(b []byte, f types.UnmarshalFunc) error {
	return f(b, &cf.ComputationTime)
}
