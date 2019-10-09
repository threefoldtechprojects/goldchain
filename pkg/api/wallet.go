package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	cftypes "github.com/nbh-digital/goldchain/extensions/custodyfees/types"
	gcmodules "github.com/nbh-digital/goldchain/modules"
)

type (
	// WalletGET contains general information about the wallet.
	WalletGET struct {
		Encrypted bool `json:"encrypted"`
		Unlocked  bool `json:"unlocked"`

		ConfirmedCoinBalance       types.Currency `json:"confirmedcoinbalance"`
		ConfirmedLockedCoinBalance types.Currency `json:"confirmedlockedcoinbalance"`
		ConfirmedCustodyFeeDebt    types.Currency `json:"confirmedcustodyfeesdebt"`
		UnconfirmedOutgoingCoins   types.Currency `json:"unconfirmedoutgoingcoins"`
		UnconfirmedIncomingCoins   types.Currency `json:"unconfirmedincomingcoins"`

		BlockStakeBalance       types.Currency `json:"blockstakebalance"`
		LockedBlockStakeBalance types.Currency `json:"lockedblockstakebalance"`

		MultiSigWallets []gcmodules.MultiSigWallet `json:"multisigwallets"`
	}

	// WalletListUnlockedGET contains the set of unspent, unlocked coin
	// and blockstake outputs owned by the wallet.
	WalletListUnlockedGET struct {
		UnlockedCoinOutputs       []UnspentCoinOutput           `json:"unlockedcoinoutputs"`
		UnlockedBlockstakeOutputs []api.UnspentBlockstakeOutput `json:"unlockedblockstakeoutputs"`
	}

	// WalletListLockedGET contains the set of unspent, locked coin and
	// blockstake outputs owned by the wallet
	WalletListLockedGET struct {
		LockedCoinOutputs       []UnspentCoinOutput           `json:"lockedcoinoutputs"`
		LockedBlockstakeOutputs []api.UnspentBlockstakeOutput `json:"lockedblockstakeoutputs"`
	}

	// UnspentCoinOutput is a coin output and its associated ID
	UnspentCoinOutput struct {
		ID       types.CoinOutputID         `json:"id"`
		Output   types.CoinOutput           `json:"coinoutput"`
		CoinInfo custodyfees.CoinOutputInfo `json:"coininfo"`
	}

	// WalletTransactionsGET contains the specified set of confirmed and
	// unconfirmed transactions.
	WalletTransactionsGET struct {
		ConfirmedTransactions   []gcmodules.ProcessedTransaction `json:"confirmedtransactions"`
		UnconfirmedTransactions []modules.ProcessedTransaction   `json:"unconfirmedtransactions"`
	}

	// WalletFundCoinsGet is the resulting object that is returned,
	// to be used by a client to fund a transaction of any type.
	WalletFundCoinsGet struct {
		CoinInputs          []types.CoinInput `json:"coininputs"`
		CustodyFeeCondition types.CoinOutput  `json:"custodyfeecondition"`
		RefundCoinOutput    *types.CoinOutput `json:"refund"`
	}
)

// RegisterWalletHTTPHandlers registers the regular handlers for all Wallet HTTP endpoints.
func RegisterWalletHTTPHandlers(router api.Router, wallet gcmodules.Wallet, requiredPassword string) {
	if wallet == nil {
		build.Critical("no wallet module given")
	}
	if router == nil {
		build.Critical("no httprouter Router given")
	}

	router.GET("/wallet", api.RequirePasswordHandler(NewWalletRootHandler(wallet), requiredPassword))
	router.GET("/wallet/blockstakestats", api.RequirePasswordHandler(api.NewWalletBlockStakeStatsHandler(wallet), requiredPassword))
	router.GET("/wallet/address", api.RequirePasswordHandler(api.NewWalletAddressHandler(wallet), requiredPassword))
	router.GET("/wallet/addresses", api.RequirePasswordHandler(api.NewWalletAddressesHandler(wallet), requiredPassword))
	router.GET("/wallet/backup", api.RequirePasswordHandler(api.NewWalletBackupHandler(wallet), requiredPassword))
	router.POST("/wallet/init", api.RequirePasswordHandler(api.NewWalletInitHandler(wallet), requiredPassword))
	router.POST("/wallet/lock", api.RequirePasswordHandler(api.NewWalletLockHandler(wallet), requiredPassword))
	router.POST("/wallet/seed", api.RequirePasswordHandler(api.NewWalletSeedHandler(wallet), requiredPassword))
	router.GET("/wallet/seeds", api.RequirePasswordHandler(api.NewWalletSeedsHandler(wallet), requiredPassword))
	router.GET("/wallet/key/:unlockhash", api.RequirePasswordHandler(api.NewWalletKeyHandler(wallet), requiredPassword))
	router.POST("/wallet/transaction", api.RequirePasswordHandler(api.NewWalletTransactionCreateHandler(wallet), requiredPassword))
	router.POST("/wallet/coins", api.RequirePasswordHandler(api.NewWalletCoinsHandler(wallet), requiredPassword))
	router.POST("/wallet/blockstakes", api.RequirePasswordHandler(api.NewWalletBlockStakesHandler(wallet), requiredPassword))
	router.GET("/wallet/transaction/:id", api.NewWalletTransactionHandler(wallet))
	router.GET("/wallet/transactions", api.NewWalletTransactionsHandler(wallet))
	router.GET("/wallet/transactions/:addr", api.NewWalletTransactionsAddrHandler(wallet))
	router.POST("/wallet/unlock", api.RequirePasswordHandler(api.NewWalletUnlockHandler(wallet), requiredPassword))
	router.GET("/wallet/unlocked", api.RequirePasswordHandler(NewWalletListUnlockedHandler(wallet), requiredPassword))
	router.GET("/wallet/locked", api.RequirePasswordHandler(NewWalletListLockedHandler(wallet), requiredPassword))
	router.POST("/wallet/create/transaction", api.RequirePasswordHandler(api.NewWalletCreateTransactionHandler(wallet), requiredPassword))
	router.POST("/wallet/sign", api.RequirePasswordHandler(api.NewWalletSignHandler(wallet), requiredPassword))
	router.GET("/wallet/publickey", api.RequirePasswordHandler(api.NewWalletGetPublicKeyHandler(wallet), requiredPassword))
	router.GET("/wallet/fund/coins", api.RequirePasswordHandler(NewWalletFundCoinsHandler(wallet), requiredPassword))
}

// NewWalletRootHandler creates a handler to handle API calls to /wallet.
func NewWalletRootHandler(wallet gcmodules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		coinBal, blockstakeBal, err := wallet.ConfirmedBalance()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		coinLockBal, blockstakeLockBal, err := wallet.ConfirmedLockedBalance()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		custodyFeesToBePaid, err := wallet.ConfirmedCustodyFeesToBePaid()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		coinsOut, coinsIn, err := wallet.UnconfirmedBalance()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		multiSigWallets, err := wallet.MultiSigWalletsWithCustodyFeeDebt()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}

		api.WriteJSON(w, WalletGET{
			Encrypted: wallet.Encrypted(),
			Unlocked:  wallet.Unlocked(),

			ConfirmedCoinBalance:       coinBal,
			ConfirmedLockedCoinBalance: coinLockBal,
			ConfirmedCustodyFeeDebt:    custodyFeesToBePaid,
			UnconfirmedOutgoingCoins:   coinsOut,
			UnconfirmedIncomingCoins:   coinsIn,

			BlockStakeBalance:       blockstakeBal,
			LockedBlockStakeBalance: blockstakeLockBal,

			MultiSigWallets: multiSigWallets,
		})
	}
}

// NewWalletListUnlockedHandler creates a handler to handle API calls to /wallet/unlocked
func NewWalletListUnlockedHandler(wallet gcmodules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		ucos, ubsos, err := wallet.UnlockedUnspendOutputsWithCustodyFeeInfo()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet/unlocked: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		ucor := []UnspentCoinOutput{}
		ubsor := []api.UnspentBlockstakeOutput{}

		for id, co := range ucos {
			ucor = append(ucor, UnspentCoinOutput{ID: id, Output: co.CoinOutput, CoinInfo: co.CoinInfo})
		}

		for id, bso := range ubsos {
			ubsor = append(ubsor, api.UnspentBlockstakeOutput{ID: id, Output: bso})
		}

		api.WriteJSON(w, WalletListUnlockedGET{
			UnlockedCoinOutputs:       ucor,
			UnlockedBlockstakeOutputs: ubsor,
		})
	}
}

// NewWalletListLockedHandler creates a handler to handle API calls to /wallet/locked
func NewWalletListLockedHandler(wallet gcmodules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		ucos, ubsos, err := wallet.LockedUnspendOutputsWithCustodyFeeInfo()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet/locked: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		ucor := []UnspentCoinOutput{}
		ubsor := []api.UnspentBlockstakeOutput{}

		for id, co := range ucos {
			ucor = append(ucor, UnspentCoinOutput{ID: id, Output: co.CoinOutput, CoinInfo: co.CoinInfo})
		}

		for id, bso := range ubsos {
			ubsor = append(ubsor, api.UnspentBlockstakeOutput{ID: id, Output: bso})
		}

		api.WriteJSON(w, WalletListLockedGET{
			LockedCoinOutputs:       ucor,
			LockedBlockstakeOutputs: ubsor,
		})
	}
}

// NewWalletTransactionsHandler creates a handler to handle API calls to /wallet/transactions.
func NewWalletTransactionsHandler(wallet gcmodules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		startheightStr, endheightStr := req.FormValue("startheight"), req.FormValue("endheight")
		if startheightStr == "" || endheightStr == "" {
			api.WriteError(w, api.Error{Message: "startheight and endheight must be provided to a /wallet/transactions call."}, http.StatusBadRequest)
			return
		}
		// Get the start and end blocks.
		start, err := strconv.Atoi(startheightStr)
		if err != nil {
			api.WriteError(w, api.Error{Message: "parsing integer value for parameter `startheight` failed: " + err.Error()}, http.StatusBadRequest)
			return
		}
		end, err := strconv.Atoi(endheightStr)
		if err != nil {
			api.WriteError(w, api.Error{Message: "parsing integer value for parameter `endheight` failed: " + err.Error()}, http.StatusBadRequest)
			return
		}
		confirmedTxns, err := wallet.TransactionsWithCustodyInfo(types.BlockHeight(start), types.BlockHeight(end))
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		unconfirmedTxns, err := wallet.UnconfirmedTransactions()
		if err != nil {
			api.WriteError(w, api.Error{Message: "error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}

		api.WriteJSON(w, WalletTransactionsGET{
			ConfirmedTransactions:   confirmedTxns,
			UnconfirmedTransactions: unconfirmedTxns,
		})
	}
}

// NewWalletFundCoinsHandler creates a handler to handle the API calls to /wallet/fund/coins?amount=.
// While it might be handy for other use cases, it is needed for 3bot registration
func NewWalletFundCoinsHandler(wallet gcmodules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		q := req.URL.Query()
		// parse the amount
		amountStr := q.Get("amount")
		if amountStr == "" || amountStr == "0" {
			api.WriteError(w, api.Error{Message: "an amount has to be specified, greater than 0"}, http.StatusBadRequest)
			return
		}
		var amount types.Currency
		err := amount.LoadString(amountStr)
		if err != nil {
			api.WriteError(w, api.Error{Message: "invalid amount given: " + err.Error()}, http.StatusBadRequest)
			return
		}

		// parse optional refund address and reuseRefundAddress from query params
		var (
			refundAddress    *types.UnlockHash
			newRefundAddress bool
		)
		refundStr := q.Get("refund")
		if refundStr != "" {
			// try as a bool
			var b bool
			n, err := fmt.Sscanf(refundStr, "%t", &b)
			if err == nil && n == 1 {
				newRefundAddress = b
			} else {
				// try as an address
				var uh types.UnlockHash
				err = uh.LoadString(refundStr)
				if err != nil {
					api.WriteError(w, api.Error{Message: fmt.Sprintf("refund query param has to be a boolean or unlockhash, %s is invalid", refundStr)}, http.StatusBadRequest)
					return
				}
				refundAddress = &uh
			}
		}

		// start a transaction and fund the requested amount
		txbuilder := wallet.StartTransaction()
		err = txbuilder.FundCoins(amount, refundAddress, !newRefundAddress)
		if err != nil {
			api.WriteError(w, api.Error{Message: "failed to fund the requested coins: " + err.Error()}, http.StatusInternalServerError)
			return
		}

		// build the dummy Txn, as to view the Txn
		txn, _ := txbuilder.View()
		// defer drop the Txn
		defer txbuilder.Drop()

		// compose the result object and validate it
		result := WalletFundCoinsGet{
			CoinInputs: txn.CoinInputs,
		}
		if len(result.CoinInputs) == 0 {
			api.WriteError(w, api.Error{Message: "no coin inputs could be generated"}, http.StatusInternalServerError)
			return
		}

		outputLength := len(txn.CoinOutputs)
		if outputLength == 0 {
			api.WriteError(w, api.Error{Message: "no coin outputs were generated, while at the very least a custody fee output was expected"}, http.StatusInternalServerError)
			return
		}
		if ct := txn.CoinOutputs[0].Condition.ConditionType(); ct != cftypes.ConditionTypeCustodyFee {
			api.WriteError(w, api.Error{Message: fmt.Sprintf("unexpected condition type %d, expected custody fee condition type as first generated coin output", ct)}, http.StatusInternalServerError)
			return
		}
		result.CustodyFeeCondition = txn.CoinOutputs[0]
		if outputLength == 2 {
			result.RefundCoinOutput = &txn.CoinOutputs[1]
		} else if outputLength > 2 {
			api.WriteError(w, api.Error{Message: "more than 2 coin outputs were generated, this is not expected"}, http.StatusInternalServerError)
			return
		}

		// all good, return the resulting object
		api.WriteJSON(w, result)
	}
}

func walletErrorToHTTPStatus(err error) int {
	if err == modules.ErrLockedWallet {
		return http.StatusForbidden
	}
	if cErr, ok := err.(types.ClientError); ok {
		return cErr.Kind.AsHTTPStatusCode()
	}
	return http.StatusInternalServerError
}
