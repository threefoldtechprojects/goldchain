package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"
)

type (
	// CoinOutputInfoGet is all coin output info that can be requested from the custody fees API about
	// a known coin output.
	CoinOutputInfoGet struct {
		CreationTime       types.Timestamp `json:"creationtime"`
		CreationValue      types.Currency  `json:"creationvalue"`
		IsCustodyFee       bool            `json:"iscustodyfee"`
		Spent              bool            `json:"spent"`
		FeeComputationTime types.Timestamp `json:"feecomputationtime"`
		CustodyFee         *types.Currency `json:"custodyfee,omitempty"`
		SpendableValue     *types.Currency `json:"spendablevalue,omitempty"`
	}
)

// NewCoinOutputInfoGetHandler creates a handler to handle the API calls to /*/custodyfees/coinoutput/:id?time=0&height=0&compute=true.
func NewCoinOutputInfoGetHandler(cs modules.ConsensusSet, plugin *custodyfees.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// load coin output ID
		var coid types.CoinOutputID
		idStr := ps.ByName("id")
		err := coid.LoadString(idStr)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: "failed to parse id param: " + err.Error()}, http.StatusBadRequest)
			return
		}

		q := req.URL.Query()

		computeFee := true
		computeFeeStr := q.Get("compute")
		if computeFeeStr != "" && (computeFeeStr == "0" || strings.ToLower(computeFeeStr) == "false") {
			computeFee = false
		}
		if !computeFee {
			info, err := plugin.GetCoinOutputInfoPreComputation(coid)
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
				return
			}
			rapi.WriteJSON(w, CoinOutputInfoGet{
				CreationTime:       info.CreationTime,
				CreationValue:      info.CreationValue,
				IsCustodyFee:       info.IsCustodyFee,
				Spent:              info.Spent,
				FeeComputationTime: info.FeeComputationTime,
				CustodyFee:         nil,
				SpendableValue:     nil,
			})
			return
		}

		// load optional time or get it from the consensus set for latest block
		var blockTime types.Timestamp
		blockTimeStr := q.Get("time")
		blockTimeIsUserDefined := blockTimeStr != ""
		if !blockTimeIsUserDefined {
			heightStr := q.Get("height")
			var height types.BlockHeight
			if heightStr != "" {
				n, err := fmt.Sscan(heightStr, &height)
				if err != nil {
					rapi.WriteError(w, rapi.Error{Message: "failed to parse time query param: " + err.Error()}, http.StatusBadRequest)
					return
				}
				if n != 1 {
					rapi.WriteError(w, rapi.Error{Message: "failed to parse time query param '" + heightStr + "'"}, http.StatusBadRequest)
					return
				}
			} else {
				height = cs.Height()
			}
			block, ok := cs.BlockAtHeight(height)
			if !ok {
				rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("failed to find block at height %d", height)}, http.StatusInternalServerError)
				return
			}
			blockTime = block.Timestamp
		} else {
			blockTime.LoadString(blockTimeStr)
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: "failed to parse time query param: " + err.Error()}, http.StatusBadRequest)
				return
			}
		}
		// get creation timestamp for coin output
		info, err := plugin.GetCoinOutputInfo(coid, blockTime)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, CoinOutputInfoGet{
			CreationTime:       info.CreationTime,
			CreationValue:      info.CreationValue,
			IsCustodyFee:       info.IsCustodyFee,
			Spent:              info.Spent,
			FeeComputationTime: info.FeeComputationTime,
			CustodyFee:         &info.CustodyFee,
			SpendableValue:     &info.SpendableValue,
		})
	}
}
