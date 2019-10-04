package api

import (
	"net/http"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"

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
	router.GET("/wallet/unlocked", api.RequirePasswordHandler(api.NewWalletListUnlockedHandler(wallet), requiredPassword))
	router.GET("/wallet/locked", api.RequirePasswordHandler(api.NewWalletListLockedHandler(wallet), requiredPassword))
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
		multiSigWallets, err := wallet.MultiSigWallets()
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

func walletErrorToHTTPStatus(err error) int {
	if err == modules.ErrLockedWallet {
		return http.StatusForbidden
	}
	if cErr, ok := err.(types.ClientError); ok {
		return cErr.Kind.AsHTTPStatusCode()
	}
	return http.StatusInternalServerError
}
