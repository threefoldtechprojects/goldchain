package api

import (
	"net/http"
	"strconv"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
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

		MultiSigWallets []modules.MultiSigWallet `json:"multisigwallets"`
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
	router.POST("/wallet/data", api.RequirePasswordHandler(api.NewWalletDataHandler(wallet), requiredPassword))
	router.GET("/wallet/transaction/:id", api.NewWalletTransactionHandler(wallet))
	router.GET("/wallet/transactions", api.NewWalletTransactionsHandler(wallet))
	router.GET("/wallet/transactions/:addr", api.NewWalletTransactionsAddrHandler(wallet))
	router.POST("/wallet/unlock", api.RequirePasswordHandler(api.NewWalletUnlockHandler(wallet), requiredPassword))
	router.GET("/wallet/unlocked", api.RequirePasswordHandler(NewWalletListUnlockedHandler(wallet), requiredPassword))
	router.GET("/wallet/locked", api.RequirePasswordHandler(NewWalletListLockedHandler(wallet), requiredPassword))
	router.POST("/wallet/create/transaction", api.RequirePasswordHandler(api.NewWalletCreateTransactionHandler(wallet), requiredPassword))
	router.POST("/wallet/sign", api.RequirePasswordHandler(api.NewWalletSignHandler(wallet), requiredPassword))
	router.GET("/wallet/publickey", api.RequirePasswordHandler(api.NewWalletGetPublicKeyHandler(wallet), requiredPassword))
	router.GET("/wallet/fund/coins", api.RequirePasswordHandler(api.NewWalletFundCoinsHandler(wallet), requiredPassword))
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

func walletErrorToHTTPStatus(err error) int {
	if err == modules.ErrLockedWallet {
		return http.StatusForbidden
	}
	if cErr, ok := err.(types.ClientError); ok {
		return cErr.Kind.AsHTTPStatusCode()
	}
	return http.StatusInternalServerError
}
