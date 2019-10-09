package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	cfexplorer "github.com/nbh-digital/goldchain/extensions/custodyfees/modules/explorer"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"
)

type (
	// ChainFactsGet is the response of the chain metrics Get explorer endpoint
	ChainFactsGet struct {
		Height types.BlockHeight `json:"height"`
		Time   types.Timestamp   `json:"time"`

		SpendableTokens       types.Currency `json:"spendabletokens"`
		SpendableLockedTokens types.Currency `json:"spendablelockedtokens"`
		TotalCustodyFeeDebt   types.Currency `json:"totalcustodyfeedebt"`

		SpentTokens     types.Currency `json:"spenttokens"`
		PaidCustodyFees types.Currency `json:"paidcustodyfees"`
	}
)

// RegisterExplorerCustodyFeesHTTPHandlers registers the default explorer HTTP handlers specific to the custodyfees package.
func RegisterExplorerCustodyFeesHTTPHandlers(router rapi.Router, cs modules.ConsensusSet, plugin *custodyfees.Plugin, explorer *cfexplorer.Explorer) {
	router.GET("/explorer/custodyfees/coinoutput/:id", NewCoinOutputInfoGetHandler(cs, plugin))
	router.GET("/explorer/custodyfees/metrics/chain", NewChainFactsGetHandler(explorer))
}

// NewChainFactsGetHandler creates a handler to handle the API calls to /explorer/custodyfees/metrics/chain.
func NewChainFactsGetHandler(explorer *cfexplorer.Explorer) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		facts, err := explorer.LatestChainFacts()
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, ChainFactsGet{
			Height: facts.Height,
			Time:   facts.Time,

			SpendableTokens:       facts.SpendableTokens,
			SpendableLockedTokens: facts.SpendableLockedTokens,
			TotalCustodyFeeDebt:   facts.TotalCustodyFeeDebt,

			SpentTokens:     facts.SpentTokens,
			PaidCustodyFees: facts.PaidCustodyFees,
		})
	}
}
