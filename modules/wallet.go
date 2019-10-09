package modules

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
)

type (
	// Wallet is an extended version of the regular Rivine Wallet
	Wallet interface {
		modules.Wallet

		// UnlockedUnspendOutputs returns all unlocked and unspend coin and blockstake outputs
		// owned by this wallet, adding custody fee information to each spent unlocked coin output.
		UnlockedUnspendOutputsWithCustodyFeeInfo() (map[types.CoinOutputID]WalletCoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error)

		// LockedUnspendOutputs returns all locked and unspend coin and blockstake outputs owned
		// by this wallet, adding custody fee information to each spent locked coin output.
		LockedUnspendOutputsWithCustodyFeeInfo() (map[types.CoinOutputID]WalletCoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error)

		// ConfirmedCustodyFeesToBePaid returns the total amount of custody fees to be paid
		// for all unlocked and locked confirmed coin outputs, in case you would spent them all.
		ConfirmedCustodyFeesToBePaid() (custodyfees types.Currency, err error)

		// Transactions returns all of the transactions that were confirmed at
		// heights [startHeight, endHeight]. Unconfirmed transactions are not
		// included.
		TransactionsWithCustodyInfo(startHeight types.BlockHeight, endHeight types.BlockHeight) ([]ProcessedTransaction, error)

		// MultiSigWalletsWithCustodyFeeDebt is the same as regular MultiSigWalletsCall but with custody fee debt included.
		MultiSigWalletsWithCustodyFeeDebt() ([]MultiSigWallet, error)
	}

	WalletCoinOutput struct {
		types.CoinOutput
		CoinInfo custodyfees.CoinOutputInfo `json:"coininfo"`
	}

	// MultiSigWallet is a collection of coin and blockstake outputs, which have the same
	// unlockhash.
	MultiSigWallet struct {
		ConfirmationBlockHeight    types.BlockHeight `json:"confirmationblockheight"`
		ConfirmationBlockTimestamp types.Timestamp   `json:"confirmationblocktime"`

		Address             types.UnlockHash           `json:"address"`
		CoinOutputIDs       []types.CoinOutputID       `json:"coinoutputids"`
		BlockStakeOutputIDs []types.BlockStakeOutputID `json:"blockstakeoutputids"`

		ConfirmedCoinBalance       types.Currency `json:"confirmedcoinbalance"`
		ConfirmedLockedCoinBalance types.Currency `json:"confirmedlockedcoinbalance"`
		ConfirmedCustodyFeeDebt    types.Currency `json:"confirmedcustodyfeedebt"`

		UnconfirmedOutgoingCoins types.Currency `json:"unconfirmedoutgoingcoins"`
		UnconfirmedIncomingCoins types.Currency `json:"unconfirmedincomingcoins"`

		ConfirmedBlockStakeBalance       types.Currency `json:"confirmedblockstakebalance"`
		ConfirmedLockedBlockStakeBalance types.Currency `json:"confirmedlockedblockstakebalance"`
		UnconfirmedOutgoingBlockStakes   types.Currency `json:"unconfirmedoutgoingblockstakes"`
		UnconfirmedIncomingBlockStakes   types.Currency `json:"unconfirmedincomingblockstakes"`

		Owners  []types.UnlockHash `json:"owners"`
		MinSigs uint64             `json:"minsigs"`
	}

	// A ProcessedInput represents funding to a transaction. The input is
	// coming from an address and going to the outputs. The fund types are
	// 'SiacoinInput', 'SiafundInput'.
	ProcessedInput struct {
		FundType types.Specifier `json:"fundtype"`
		// WalletAddress indicates it's an address owned by this wallet
		WalletAddress  bool                        `json:"walletaddress"`
		RelatedAddress types.UnlockHash            `json:"relatedaddress"`
		Value          types.Currency              `json:"value"`
		ParentOutputID types.OutputID              `json:"parentid"`
		CoinInfo       *custodyfees.CoinOutputInfo `json:"coininfo,omitempty"`
	}

	// A ProcessedOutput is a coin output that appears in a transaction.
	// Some outputs mature immediately, some are delayed.
	//
	// Fund type can either be 'CoinOutput', 'BlockStakeOutput'
	// or 'MinerFee'. All outputs except the miner fee create
	// outputs accessible to an address. Miner fees are not spendable, and
	// instead contribute to the block subsidy.
	//
	// MaturityHeight indicates at what block height the output becomes
	// available. CoinInputs and BlockStakeInputs become available immediately.
	// MinerPayouts become available after 144 confirmations.
	ProcessedOutput struct {
		FundType       types.Specifier   `json:"fundtype"`
		MaturityHeight types.BlockHeight `json:"maturityheight"`
		// WalletAddress indicates it's an address owned by this wallet
		WalletAddress  bool                        `json:"walletaddress"`
		RelatedAddress types.UnlockHash            `json:"relatedaddress"`
		Value          types.Currency              `json:"value"`
		OutputID       types.OutputID              `json:"id"`
		CoinInfo       *custodyfees.CoinOutputInfo `json:"coininfo,omitempty"`
	}

	// A ProcessedTransaction is a transaction that has been processed into
	// explicit inputs and outputs and tagged with some header data such as
	// confirmation height + timestamp.
	//
	// Because of the block subsidy, a block is considered as a transaction.
	// Since there is technically no transaction id for the block subsidy, the
	// block id is used instead.
	ProcessedTransaction struct {
		Transaction           types.Transaction   `json:"transaction"`
		TransactionID         types.TransactionID `json:"transactionid"`
		ConfirmationHeight    types.BlockHeight   `json:"confirmationheight"`
		ConfirmationTimestamp types.Timestamp     `json:"confirmationtimestamp"`

		Inputs  []ProcessedInput  `json:"inputs"`
		Outputs []ProcessedOutput `json:"outputs"`
	}

	// WalletProcessedInput is similar to a ProcessedInput with with only static values.
	WalletProcessedInput struct {
		FundType types.Specifier `json:"fundtype"`
		// WalletAddress indicates it's an address owned by this wallet
		WalletAddress  bool             `json:"walletaddress"`
		RelatedAddress types.UnlockHash `json:"relatedaddress"`
		Value          types.Currency   `json:"value"`
		ParentOutputID types.OutputID   `json:"parentid"`
	}

	// WalletProcessedOutput is similar to a ProcssedOutput with with only static values.
	WalletProcessedOutput struct {
		FundType       types.Specifier   `json:"fundtype"`
		MaturityHeight types.BlockHeight `json:"maturityheight"`
		// WalletAddress indicates it's an address owned by this wallet
		WalletAddress  bool             `json:"walletaddress"`
		RelatedAddress types.UnlockHash `json:"relatedaddress"`
		Value          types.Currency   `json:"value"`
		OutputID       types.OutputID   `json:"id"`
	}

	// WalletProcessedTransaction is similar to a regular ProcessedTransaction but with only static values.
	WalletProcessedTransaction struct {
		Transaction           types.Transaction   `json:"transaction"`
		TransactionID         types.TransactionID `json:"transactionid"`
		ConfirmationHeight    types.BlockHeight   `json:"confirmationheight"`
		ConfirmationTimestamp types.Timestamp     `json:"confirmationtimestamp"`

		Inputs  []WalletProcessedInput  `json:"inputs"`
		Outputs []WalletProcessedOutput `json:"outputs"`
	}
)

func (gpi *ProcessedInput) AsRivineProcessedInput() modules.ProcessedInput {
	rpi := modules.ProcessedInput{
		FundType:       gpi.FundType,
		WalletAddress:  gpi.WalletAddress,
		RelatedAddress: gpi.RelatedAddress,
		Value:          gpi.Value,
	}
	if gpi.CoinInfo != nil && !gpi.CoinInfo.SpendableValue.IsZero() {
		rpi.Value = gpi.CoinInfo.SpendableValue
	}
	return rpi
}

func (gpo *ProcessedOutput) AsRivineProcessedOutput() modules.ProcessedOutput {
	rpo := modules.ProcessedOutput{
		FundType:       gpo.FundType,
		MaturityHeight: gpo.MaturityHeight,
		WalletAddress:  gpo.WalletAddress,
		RelatedAddress: gpo.RelatedAddress,
		Value:          gpo.Value,
	}
	if gpo.CoinInfo != nil && !gpo.CoinInfo.SpendableValue.IsZero() {
		rpo.Value = gpo.CoinInfo.SpendableValue
	}
	return rpo
}

func (gpt *ProcessedTransaction) AsRivineProcessedTransaction() modules.ProcessedTransaction {
	rpt := modules.ProcessedTransaction{
		Transaction:           gpt.Transaction,
		TransactionID:         gpt.TransactionID,
		ConfirmationHeight:    gpt.ConfirmationHeight,
		ConfirmationTimestamp: gpt.ConfirmationTimestamp,

		Inputs:  make([]modules.ProcessedInput, 0, len(gpt.Inputs)),
		Outputs: make([]modules.ProcessedOutput, 0, len(gpt.Outputs)),
	}
	for idx := range gpt.Inputs {
		rpt.Inputs = append(rpt.Inputs, gpt.Inputs[idx].AsRivineProcessedInput())
	}
	for idx := range gpt.Outputs {
		rpt.Outputs = append(rpt.Outputs, gpt.Outputs[idx].AsRivineProcessedOutput())
	}
	return rpt
}

func (wpi *WalletProcessedInput) AsRivineProcessedInput() modules.ProcessedInput {
	return modules.ProcessedInput{
		FundType:       wpi.FundType,
		WalletAddress:  wpi.WalletAddress,
		RelatedAddress: wpi.RelatedAddress,
		Value:          wpi.Value,
	}
}

func (wpo *WalletProcessedOutput) AsRivineProcessedOutput() modules.ProcessedOutput {
	return modules.ProcessedOutput{
		FundType:       wpo.FundType,
		MaturityHeight: wpo.MaturityHeight,
		WalletAddress:  wpo.WalletAddress,
		RelatedAddress: wpo.RelatedAddress,
		Value:          wpo.Value,
	}
}

func (wpt *WalletProcessedTransaction) AsRivineProcessedTransaction() modules.ProcessedTransaction {
	rpt := modules.ProcessedTransaction{
		Transaction:           wpt.Transaction,
		TransactionID:         wpt.TransactionID,
		ConfirmationHeight:    wpt.ConfirmationHeight,
		ConfirmationTimestamp: wpt.ConfirmationTimestamp,

		Inputs:  make([]modules.ProcessedInput, 0, len(wpt.Inputs)),
		Outputs: make([]modules.ProcessedOutput, 0, len(wpt.Outputs)),
	}
	for idx := range wpt.Inputs {
		rpt.Inputs = append(rpt.Inputs, wpt.Inputs[idx].AsRivineProcessedInput())
	}
	for idx := range wpt.Outputs {
		rpt.Outputs = append(rpt.Outputs, wpt.Outputs[idx].AsRivineProcessedOutput())
	}
	return rpt
}

func (wpi *WalletProcessedInput) AsProcessedInput(view custodyfees.CoinOutputInfoView, blockTime types.Timestamp) (ProcessedInput, error) {
	pi := ProcessedInput{
		FundType:       wpi.FundType,
		WalletAddress:  wpi.WalletAddress,
		RelatedAddress: wpi.RelatedAddress,
		Value:          wpi.Value,
		ParentOutputID: wpi.ParentOutputID,
	}
	if pi.FundType != types.SpecifierBlockStakeInput {
		info, err := view.GetCoinOutputInfo(types.CoinOutputID(wpi.ParentOutputID), blockTime)
		if err != nil {
			return ProcessedInput{}, err
		}
		pi.CoinInfo = &info
	}
	return pi, nil
}

func (wpo *WalletProcessedOutput) AsProcessedOutput(view custodyfees.CoinOutputInfoView, blockTime types.Timestamp) (ProcessedOutput, error) {
	po := ProcessedOutput{
		FundType:       wpo.FundType,
		MaturityHeight: wpo.MaturityHeight,
		WalletAddress:  wpo.WalletAddress,
		RelatedAddress: wpo.RelatedAddress,
		Value:          wpo.Value,
		OutputID:       wpo.OutputID,
	}
	if po.FundType != types.SpecifierBlockStakeOutput {
		info, err := view.GetCoinOutputInfo(types.CoinOutputID(wpo.OutputID), blockTime)
		if err != nil {
			return ProcessedOutput{}, err
		}
		po.CoinInfo = &info
	}
	return po, nil
}

func (wpt *WalletProcessedTransaction) AsProcessedTransaction(view custodyfees.CoinOutputInfoView) (ProcessedTransaction, error) {
	rpt := ProcessedTransaction{
		Transaction:           wpt.Transaction,
		TransactionID:         wpt.TransactionID,
		ConfirmationHeight:    wpt.ConfirmationHeight,
		ConfirmationTimestamp: wpt.ConfirmationTimestamp,

		Inputs:  make([]ProcessedInput, 0, len(wpt.Inputs)),
		Outputs: make([]ProcessedOutput, 0, len(wpt.Outputs)),
	}
	var (
		err error
		pi  ProcessedInput
		po  ProcessedOutput
	)
	for idx := range wpt.Inputs {
		pi, err = wpt.Inputs[idx].AsProcessedInput(view, wpt.ConfirmationTimestamp)
		if err != nil {
			return ProcessedTransaction{}, err
		}
		rpt.Inputs = append(rpt.Inputs, pi)
	}
	for idx := range wpt.Outputs {
		po, err = wpt.Outputs[idx].AsProcessedOutput(view, wpt.ConfirmationTimestamp)
		if err != nil {
			return ProcessedTransaction{}, err
		}
		rpt.Outputs = append(rpt.Outputs, po)
	}
	return rpt, nil
}
