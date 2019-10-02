package api

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"
)

// CoinOutputGetAge contains the requested age of a coin output,
// for a given coin output ID and optional block time.
type CoinOutputGetAge struct {
	Age types.Timestamp `json:"age"`
}

// CoinOutputGetCustodyFee contains the requested fee and age of a coin output,
// for a given coin output ID and optional block time.
type CoinOutputGetCustodyFee struct {
	Value types.Currency  `json:"value"` // amount still spendable
	Fee   types.Currency  `json:"fee"`
	Age   types.Timestamp `json:"age"`
}

// NewCoinOutputGetAgeHandler creates a handler to handle the API calls to /*/custodyfees/coinoutput/age/:id?time=0.
func NewCoinOutputGetAgeHandler(cs modules.ConsensusSet, plugin *custodyfees.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		coid, blockTime, blockTimeIsUserDefined, ok := getCoinOutputIDAndTimeFromParams(cs, ps, w)
		if !ok {
			return
		}
		age, ok := getAgeOfCoinOutput(plugin, coid, blockTime, blockTimeIsUserDefined, w)
		if !ok {
			return
		}
		rapi.WriteJSON(w, CoinOutputGetAge{
			Age: age,
		})
	}
}

// NewCoinOutputGetCustodyFeeHandler creates a handler to handle the API calls to /*/custodyfees/coinoutput/fee/:id?time=0.
func NewCoinOutputGetCustodyFeeHandler(cs modules.ConsensusSet, plugin *custodyfees.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		coid, blockTime, blockTimeIsUserDefined, ok := getCoinOutputIDAndTimeFromParams(cs, ps, w)
		if !ok {
			return
		}
		age, ok := getAgeOfCoinOutput(plugin, coid, blockTime, blockTimeIsUserDefined, w)
		if !ok {
			return
		}
		co, err := cs.GetCoinOutput(coid)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: "failed to look up coin output in consensus set: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		_, fee := custodyfees.AmountCustodyFeePairAfterXSeconds(co.Value, age)
		rapi.WriteJSON(w, CoinOutputGetCustodyFee{
			Value: co.Value.Sub(fee),
			Fee:   fee,
			Age:   age,
		})
	}
}

func getCoinOutputIDAndTimeFromParams(cs modules.ConsensusSet, ps httprouter.Params, w http.ResponseWriter) (coid types.CoinOutputID, blockTime types.Timestamp, blockTimeIsUserDefined bool, ok bool) {
	// load coin output ID
	idStr := ps.ByName("id")
	err := coid.LoadString(idStr)
	if err != nil {
		rapi.WriteError(w, rapi.Error{Message: "failed to parse id param: " + err.Error()}, http.StatusBadRequest)
		return
	}

	// load optional time or get it from the consensus set for latest block
	blockTimeStr := ps.ByName("time")
	blockTimeIsUserDefined = blockTimeStr != ""
	if !blockTimeIsUserDefined {
		height := cs.Height()
		var block types.Block
		block, ok = cs.BlockAtHeight(height)
		if !ok {
			rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("failed to find block at current height %d", height)}, http.StatusInternalServerError)
			return
		}
		blockTime = block.Timestamp
	} else {
		blockTime.LoadString(blockTimeStr)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: "failed to parse time param: " + err.Error()}, http.StatusBadRequest)
			return
		}
	}
	ok = true
	return
}

func getAgeOfCoinOutput(plugin *custodyfees.Plugin, id types.CoinOutputID, blockTime types.Timestamp, blockTimeIsUsedDefined bool, w http.ResponseWriter) (types.Timestamp, bool) {
	// get creation timestamp for coin output
	ts, err := plugin.GetCoinOutputCreationTime(id)
	if err != nil {
		rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
		return 0, false
	}

	// calculate age of coin output and return it
	if ts > blockTime {
		status := http.StatusBadRequest
		if !blockTimeIsUsedDefined {
			status = http.StatusInternalServerError
		}
		rapi.WriteError(w, rapi.Error{
			Message: fmt.Sprintf("invalid coin output creation time is in future compare to used block timestamp: %d > %d", ts, blockTime)},
			status)
		return 0, false
	}
	return blockTime - ts, true
}
