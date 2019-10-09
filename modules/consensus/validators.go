package consensus

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/consensus"
	"github.com/threefoldtech/rivine/types"

	cftypes "github.com/nbh-digital/goldchain/extensions/custodyfees/types"
)

func GetTestnetTransactionValidators() []modules.TransactionValidationFunction {
	return getTransactionValidators()
}

func GetTestnetTransactionVersionMappedValidators() map[types.TransactionVersion][]modules.TransactionValidationFunction {
	return getTransactionVersionMappedValidators()
}

func GetDevnetTransactionValidators() []modules.TransactionValidationFunction {
	return getTransactionValidators()
}

func GetDevnetTransactionVersionMappedValidators() map[types.TransactionVersion][]modules.TransactionValidationFunction {
	return getTransactionVersionMappedValidators()
}

// ValidateCoinOutputsAreValid is a validator function that checks if all coin outputs are standard,
// meaning their condition is considered standard (== known) and their (coin) value is individually greater than zero,
// the exception is that Custody Fees are allowed to have a value equal to zero.
func ValidateCoinOutputsAreValid(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var err error
	for _, co := range tx.CoinOutputs {
		if co.Value.IsZero() && co.Condition.ConditionType() != cftypes.ConditionTypeCustodyFee {
			return types.ErrZeroOutput
		}
		err = co.Condition.IsStandardCondition(ctx.ValidationContext)
		if err != nil {
			return err
		}
	}
	return nil
}

func getTransactionVersionMappedValidators() map[types.TransactionVersion][]modules.TransactionValidationFunction {
	return map[types.TransactionVersion][]modules.TransactionValidationFunction{
		types.TransactionVersionZero: {
			consensus.ValidateInvalidByDefault,
		},
		types.TransactionVersionOne: {
			consensus.ValidateCoinOutputsAreBalanced,
			consensus.ValidateBlockStakeOutputsAreBalanced,
			consensus.ValidateMinerFeeIsPresent,
		},
	}
}

func getTransactionValidators() []modules.TransactionValidationFunction {
	return []modules.TransactionValidationFunction{
		consensus.ValidateTransactionFitsInABlock,
		consensus.ValidateTransactionArbitraryData,
		consensus.ValidateCoinInputsAreValid,
		ValidateCoinOutputsAreValid,
		consensus.ValidateBlockStakeInputsAreValid,
		consensus.ValidateBlockStakeOutputsAreValid,
		consensus.ValidateMinerFeesAreValid,
		consensus.ValidateDoubleCoinSpends,
		consensus.ValidateDoubleBlockStakeSpends,
		consensus.ValidateCoinInputsAreFulfilled,
		consensus.ValidateBlockStakeInputsAreFulfilled,
	}
}
