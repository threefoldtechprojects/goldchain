package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
	rtypes "github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	cftypes "github.com/nbh-digital/goldchain/extensions/custodyfees/types"
)

// Primitive HTTP Explorer API Types
type (
	// ExplorerCoinOutput is the same a regular rivine.api.ExplorerCoinOutput,
	// but with the addition that the coin output custody fee information is also attached.
	ExplorerCoinOutput struct {
		rapi.ExplorerCoinOutput
		Custody CustodyFeeInfo `json:"custody"`
	}

	// ExplorerBlock is a block with some extra information such as the id and
	// height. This information is provided for programs that may not be
	// complex enough to compute the ID on their own.
	ExplorerBlock struct {
		MinerPayoutIDs         []rtypes.CoinOutputID `json:"minerpayoutids"`
		MinerPayoutCustodyFees []CustodyFeeInfo      `json:"minerpayoutcustodyfees"`
		Transactions           []ExplorerTransaction `json:"transactions"`
		RawBlock               rtypes.Block          `json:"rawblock"`

		modules.BlockFacts
	}

	// ExplorerTransaction is a transcation with some extra information such as
	// the parent block. This information is provided for programs that may not
	// be complex enough to compute the extra information on their own.
	ExplorerTransaction struct {
		ID             rtypes.TransactionID `json:"id"`
		Height         rtypes.BlockHeight   `json:"height"`
		Parent         rtypes.BlockID       `json:"parent"`
		RawTransaction rtypes.Transaction   `json:"rawtransaction"`
		Timestamp      rtypes.Timestamp     `json:"timestamp"`
		Order          int                  `json:"order"`

		CoinInputOutputs             []ExplorerCoinOutput            `json:"coininputoutputs"` // the outputs being spent
		CoinOutputIDs                []rtypes.CoinOutputID           `json:"coinoutputids"`
		CoinOutputUnlockHashes       []rtypes.UnlockHash             `json:"coinoutputunlockhashes"`
		CoinOutputCustodyFees        []CustodyFeeInfo                `json:"coinoutputcustodyfees"`
		BlockStakeInputOutputs       []rapi.ExplorerBlockStakeOutput `json:"blockstakeinputoutputs"` // the outputs being spent
		BlockStakeOutputIDs          []rtypes.BlockStakeOutputID     `json:"blockstakeoutputids"`
		BlockStakeOutputUnlockHashes []rtypes.UnlockHash             `json:"blockstakeunlockhashes"`

		Unconfirmed bool `json:"unconfirmed"`
	}

	// CustodyFeeInfo contains the fee for a certain coin output as well as the age at the time of fee calculation.
	CustodyFeeInfo struct {
		CreationTime       rtypes.Timestamp `json:"creationtime"`
		CreationValue      rtypes.Currency  `json:"creationvalue"`
		IsCustodyFee       bool             `json:"iscustodyfee"`
		Spent              bool             `json:"spent"`
		FeeComputationTime rtypes.Timestamp `json:"feecomputationtime"`
		CustodyFee         rtypes.Currency  `json:"custodyfee"`
		SpendableValue     rtypes.Currency  `json:"spendablevalue"`
	}
)

// HTTP response objects
type (
	// ExplorerBlockGET is the object returned by a GET request to
	// /explorer/block.
	ExplorerBlockGET struct {
		Block ExplorerBlock `json:"block"`
	}

	// ExplorerHashGET is the object returned as a response to a GET request to
	// /explorer/hash. The HashType will indicate whether the hash corresponds
	// to a block id, a transaction id, a siacoin output id, a file contract
	// id, or a siafund output id. In the case of a block id, 'Block' will be
	// filled out and all the rest of the fields will be blank. In the case of
	// a transaction id, 'Transaction' will be filled out and all the rest of
	// the fields will be blank. For everything else, 'Transactions' and
	// 'Blocks' will/may be filled out and everything else will be blank.
	ExplorerHashGET struct {
		HashType          string                `json:"hashtype"`
		Block             ExplorerBlock         `json:"block"`
		Blocks            []ExplorerBlock       `json:"blocks"`
		Transaction       ExplorerTransaction   `json:"transaction"`
		Transactions      []ExplorerTransaction `json:"transactions"`
		MultiSigAddresses []rtypes.UnlockHash   `json:"multisigaddresses"`
		Unconfirmed       bool                  `json:"unconfirmed"`
	}
)

// RegisterExplorerHTTPHandlers registers the (tfchain-specific) handlers for all Explorer HTTP endpoints.
func RegisterExplorerHTTPHandlers(router rapi.Router, cs modules.ConsensusSet, explorer modules.Explorer, tpool modules.TransactionPool, plugin *custodyfees.Plugin) {
	if cs == nil {
		panic("no ConsensusSet API given")
	}
	if explorer == nil {
		panic("no Explorer API given")
	}
	if plugin == nil {
		panic("no CustodyFees plugin given")
	}
	if router == nil {
		panic("no router given")
	}

	// rivine endpoints

	router.GET("/explorer", rapi.NewExplorerRootHandler(explorer))
	router.GET("/explorer/stats/history", rapi.NewExplorerHistoryStatsHandler(explorer))
	router.GET("/explorer/stats/range", rapi.NewExplorerRangeStatsHandler(explorer))
	router.GET("/explorer/constants", rapi.NewExplorerConstantsHandler(explorer))
	router.GET("/explorer/downloader/status", rapi.NewConsensusRootHandler(cs))

	// goldchain-overwritten endpoints (Custody Fees Info Support)

	router.GET("/explorer/blocks/:height", NewExplorerBlocksHandler(cs, explorer, plugin))
	router.GET("/explorer/hashes/:hash", NewExplorerHashHandler(explorer, tpool, cs, plugin))
}

// NewExplorerBlocksHandler creates a handler to handle API calls to /explorer/blocks/:height.
func NewExplorerBlocksHandler(cs modules.ConsensusSet, explorer modules.Explorer, plugin *custodyfees.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Parse the height that's being requested.
		var height rtypes.BlockHeight
		_, err := fmt.Sscan(ps.ByName("height"), &height)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusBadRequest)
			return
		}

		// get the current block info
		currentHeight := cs.Height()
		currentBlock, exists := cs.BlockAtHeight(currentHeight)
		if !exists {
			rapi.WriteError(w, rapi.Error{Message: "failed to find latest block info"}, http.StatusInternalServerError)
			return
		}

		// Fetch and return the explorer block.
		block, exists := cs.BlockAtHeight(height)
		if !exists {
			rapi.WriteError(w, rapi.Error{Message: "no block found at input height in call to /explorer/block"}, http.StatusBadRequest)
			return
		}
		rapi.WriteJSON(w, ExplorerBlockGET{
			Block: buildExplorerBlock(explorer, plugin, currentBlock.Timestamp, height, block),
		})
	}
}

// NewExplorerHashHandler creates a handler to handle GET requests to /explorer/hash/:hash.
func NewExplorerHashHandler(explorer modules.Explorer, tpool modules.TransactionPool, cs modules.ConsensusSet, plugin *custodyfees.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// get the current block info
		currentHeight := cs.Height()
		currentBlock, exists := cs.BlockAtHeight(currentHeight)
		if !exists {
			rapi.WriteError(w, rapi.Error{Message: "failed to find latest block info"}, http.StatusInternalServerError)
			return
		}

		// Scan the hash as a hash. If that fails, try scanning the hash as an
		// address.
		hash, err := rapi.ScanHash(ps.ByName("hash"))
		if err != nil {
			addr, err := rapi.ScanAddress(ps.ByName("hash"))
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusBadRequest)
				return
			}

			// Try the hash as an unlock hash. Unlock hash is checked last because
			// unlock hashes do not have collision-free guarantees. Someone can create
			// an unlock hash that collides with another object id. They will not be
			// able to use the unlock hash, but they can disrupt the explorer. This is
			// handled by checking the unlock hash last. Anyone intentionally creating
			// a colliding unlock hash (such a collision can only happen if done
			// intentionally) will be unable to find their unlock hash in the
			// blockchain through the explorer hash lookup.
			var (
				txns   []ExplorerTransaction
				blocks []ExplorerBlock
			)
			if txids := explorer.UnlockHash(addr); len(txids) != 0 {
				// parse the optional filters for the unlockhash request
				var filters rapi.TransactionSetFilters
				if str := req.FormValue("minheight"); str != "" {
					n, err := strconv.ParseUint(str, 10, 64)
					if err != nil {
						rapi.WriteError(w, rapi.Error{Message: "invalid minheight filter: " + err.Error()}, http.StatusBadRequest)
						return
					}
					filters.MinBlockHeight = rtypes.BlockHeight(n)
				}
				// build the transaction set for all transactions for the given unlock hash,
				// taking into account the given filters
				txns, blocks = buildTransactionSet(explorer, plugin, currentBlock.Timestamp, txids, filters)
			}
			txns = append(txns, getUnconfirmedTransactions(explorer, plugin, currentBlock.Timestamp, tpool, addr)...)
			multiSigAddresses := explorer.MultiSigAddresses(addr)
			if len(txns) != 0 || len(blocks) != 0 || len(multiSigAddresses) != 0 {
				// Sort transactions by height
				sort.Sort(explorerTransactionsByHeight(txns))

				rapi.WriteJSON(w, ExplorerHashGET{
					HashType:          rapi.HashTypeUnlockHashStr,
					Blocks:            blocks,
					Transactions:      txns,
					MultiSigAddresses: multiSigAddresses,
				})
				return
			}

			// Hash not found, return an error.
			rapi.WriteError(w, rapi.Error{Message: "no transactions or blocks found for given address"}, http.StatusNoContent)
			return
		}

		// TODO: lookups on the zero hash are too expensive to allow. Need a
		// better way to handle this case.
		if hash == (crypto.Hash{}) {
			rapi.WriteError(w, rapi.Error{Message: "can't lookup the empty unlock hash"}, http.StatusBadRequest)
			return
		}

		// Try the hash as a block id.
		block, height, exists := explorer.Block(rtypes.BlockID(hash))
		if exists {
			rapi.WriteJSON(w, ExplorerHashGET{
				HashType: rapi.HashTypeBlockIDStr,
				Block:    buildExplorerBlock(explorer, plugin, currentBlock.Timestamp, height, block),
			})
			return
		}

		// Try the hash as a transaction id.
		block, height, exists = explorer.Transaction(rtypes.TransactionID(hash))
		if exists {
			var txn rtypes.Transaction
			for _, t := range block.Transactions {
				if t.ID() == rtypes.TransactionID(hash) {
					txn = t
				}
			}
			rapi.WriteJSON(w, ExplorerHashGET{
				HashType:    rapi.HashTypeTransactionIDStr,
				Transaction: buildExplorerTransaction(explorer, plugin, currentBlock.Timestamp, height, block, txn),
			})
			return
		}

		// Try the hash as a siacoin output id.
		txids := explorer.CoinOutputID(rtypes.CoinOutputID(hash))
		if len(txids) != 0 {
			txns, blocks := buildTransactionSet(explorer, plugin, currentBlock.Timestamp, txids, rapi.TransactionSetFilters{})
			// Sort transactions by height
			sort.Sort(explorerTransactionsByHeight(txns))

			rapi.WriteJSON(w, ExplorerHashGET{
				HashType:     rapi.HashTypeCoinOutputIDStr,
				Blocks:       blocks,
				Transactions: txns,
			})
			return
		}

		// Try the hash as a siafund output id.
		txids = explorer.BlockStakeOutputID(rtypes.BlockStakeOutputID(hash))
		if len(txids) != 0 {
			txns, blocks := buildTransactionSet(explorer, plugin, currentBlock.Timestamp, txids, rapi.TransactionSetFilters{})
			// Sort transactions by height
			sort.Sort(explorerTransactionsByHeight(txns))

			rapi.WriteJSON(w, ExplorerHashGET{
				HashType:     rapi.HashTypeBlockStakeOutputIDStr,
				Blocks:       blocks,
				Transactions: txns,
			})
			return
		}

		// if the transaction pool is available, try to use it
		if tpool != nil {
			// Try the hash as a transactionID in the transaction pool
			txn, err := tpool.Transaction(rtypes.TransactionID(hash))
			if err == nil {
				rapi.WriteJSON(w, ExplorerHashGET{
					HashType:    rapi.HashTypeTransactionIDStr,
					Transaction: buildExplorerTransaction(explorer, plugin, currentBlock.Timestamp, 0, rtypes.Block{}, txn),
					Unconfirmed: true,
				})
				return
			}
			if err != modules.ErrTransactionNotFound {
				rapi.WriteError(w, rapi.Error{
					Message: "error during call to /explorer/hash: failed to get txn from transaction pool: " + err.Error()},
					http.StatusInternalServerError)
				return
			}
		}

		// Hash not found, return an error.
		rapi.WriteError(w, rapi.Error{Message: "unrecognized hash used as input to /explorer/hash"}, http.StatusBadRequest)
	}
}

// buildTransactionSet returns the blocks and transactions that are associated
// with a set of transaction ids.
func buildTransactionSet(explorer modules.Explorer, plugin *custodyfees.Plugin, chainTime rtypes.Timestamp, txids []rtypes.TransactionID, filters rapi.TransactionSetFilters) (txns []ExplorerTransaction, blocks []ExplorerBlock) {
	for _, txid := range txids {
		// Get the block containing the transaction - in the case of miner
		// payouts, the block might be the transaction.
		block, height, exists := explorer.Transaction(txid)
		if !exists {
			build.Severe("explorer pointing to nonexistent txn")
		}

		// ensure the height is within the minimum range
		if height < filters.MinBlockHeight {
			continue // skip this block
		}

		// Check if the block is the transaction.
		if rtypes.TransactionID(block.ID()) == txid {
			blocks = append(blocks, buildExplorerBlock(explorer, plugin, chainTime, height, block))
		} else {
			// Find the transaction within the block with the correct id.
			for _, t := range block.Transactions {
				if t.ID() == txid {
					txns = append(txns, buildExplorerTransaction(explorer, plugin, chainTime, height, block, t))
					break
				}
			}
		}
	}
	return txns, blocks
}

// buildExplorerBlock takes a block and its height and uses it to construct an explorer block.
func buildExplorerBlock(explorer modules.Explorer, plugin *custodyfees.Plugin, chainTime rtypes.Timestamp, height rtypes.BlockHeight, block rtypes.Block) ExplorerBlock {
	var (
		mpoids      []rtypes.CoinOutputID
		custodyFees []CustodyFeeInfo
	)
	for i := range block.MinerPayouts {
		mpoid := block.MinerPayoutID(uint64(i))
		mpoids = append(mpoids, mpoid)
		feeInfo, err := getCoinOutputCustodyFeeInfo(plugin, chainTime, rtypes.CoinOutputID(mpoid))
		if err != nil {
			build.Severe("error while fetching custody info for coin output", err)
		}
		custodyFees = append(custodyFees, feeInfo)
	}

	var etxns []ExplorerTransaction
	for _, txn := range block.Transactions {
		etxns = append(etxns, buildExplorerTransaction(explorer, plugin, chainTime, height, block, txn))
	}

	facts, exists := explorer.BlockFacts(height)
	if !exists {
		build.Severe("incorrect request to buildExplorerBlock - block does not exist")
	}

	return ExplorerBlock{
		MinerPayoutIDs:         mpoids,
		MinerPayoutCustodyFees: custodyFees,
		Transactions:           etxns,
		RawBlock:               block,

		BlockFacts: facts,
	}
}

// getUnconfirmedTransactions returns a list of all transactions which are unconfirmed and related to the given unlock hash from the transactionpool
func getUnconfirmedTransactions(explorer modules.Explorer, plugin *custodyfees.Plugin, chainTime rtypes.Timestamp, tpool modules.TransactionPool, addr rtypes.UnlockHash) []ExplorerTransaction {
	if tpool == nil {
		return nil
	}
	relatedTxns := []rtypes.Transaction{}
	unconfirmedTxns := tpool.TransactionList()
	// make a list of potential unspend coin outputs
	potentiallySpentCoinOutputs := map[rtypes.CoinOutputID]rtypes.CoinOutput{}
	for _, txn := range unconfirmedTxns {
		for idx, sco := range txn.CoinOutputs {
			potentiallySpentCoinOutputs[txn.CoinOutputID(uint64(idx))] = sco
		}
	}
	// go through all unconfirmed transactions
unconfirmedTxsLoop:
	for _, txn := range unconfirmedTxns {
		// Check if any coin output is related to the hash we currently have
		for _, co := range txn.CoinOutputs {
			if co.Condition.UnlockHash() == addr {
				relatedTxns = append(relatedTxns, txn)
				continue unconfirmedTxsLoop
			}
		}
		// Check if any blockstake output is related
		for _, bso := range txn.BlockStakeOutputs {
			if bso.Condition.UnlockHash() == addr {
				relatedTxns = append(relatedTxns, txn)
				continue unconfirmedTxsLoop
			}
		}
		// Check the coin inputs
		for _, ci := range txn.CoinInputs {
			// check if related to an unconfirmed coin output
			if sco, ok := potentiallySpentCoinOutputs[ci.ParentID]; ok && sco.Condition.UnlockHash() == addr {
				// mark related, add tx and stop this coin input loop
				relatedTxns = append(relatedTxns, txn)
				continue unconfirmedTxsLoop
			}
			// check if related to a confirmed coin output
			co, _ := explorer.CoinOutput(ci.ParentID)
			if co.Condition.UnlockHash() == addr {
				relatedTxns = append(relatedTxns, txn)
				continue unconfirmedTxsLoop
			}
		}
		// Check blockstake inputs
		for _, bsi := range txn.BlockStakeInputs {
			bsi, _ := explorer.BlockStakeOutput(bsi.ParentID)
			if bsi.Condition.UnlockHash() == addr {
				relatedTxns = append(relatedTxns, txn)
				continue unconfirmedTxsLoop
			}
		}
	}
	explorerTxns := make([]ExplorerTransaction, len(relatedTxns))
	for i := range relatedTxns {
		relatedTxn := relatedTxns[i]
		spentCoinOutputs := map[rtypes.CoinOutputID]rtypes.CoinOutput{}
		for _, sci := range relatedTxn.CoinInputs {
			// add unconfirmed coin output
			if sco, ok := potentiallySpentCoinOutputs[sci.ParentID]; ok {
				spentCoinOutputs[sci.ParentID] = sco
				continue
			}
			// add confirmed coin output
			sco, exists := explorer.CoinOutput(sci.ParentID)
			if !exists {
				build.Critical("could not find corresponding coin output")
			}
			spentCoinOutputs[sci.ParentID] = sco
		}
		explorerTxns[i] = buildExplorerTransactionWithMappedCoinOutputs(explorer, plugin, chainTime, 0, rtypes.Block{}, relatedTxn, spentCoinOutputs, false)
		explorerTxns[i].Unconfirmed = true
	}
	return explorerTxns
}

// buildExplorerTransaction takes a transaction and the height + id of the
// block it appears in an uses that to build an explorer transaction.
func buildExplorerTransaction(explorer modules.Explorer, plugin *custodyfees.Plugin, chainTime rtypes.Timestamp, height rtypes.BlockHeight, block rtypes.Block, txn rtypes.Transaction) (et ExplorerTransaction) {
	spentCoinOutputs := map[rtypes.CoinOutputID]rtypes.CoinOutput{}
	for _, sci := range txn.CoinInputs {
		sco, exists := explorer.CoinOutput(sci.ParentID)
		if !exists {
			build.Severe("could not find corresponding coin output")
		}
		spentCoinOutputs[sci.ParentID] = sco
	}
	return buildExplorerTransactionWithMappedCoinOutputs(explorer, plugin, chainTime, height, block, txn, spentCoinOutputs, true)
}

func buildExplorerTransactionWithMappedCoinOutputs(explorer modules.Explorer, plugin *custodyfees.Plugin, chainTime rtypes.Timestamp, height rtypes.BlockHeight, block rtypes.Block, txn rtypes.Transaction, spentCoinOutputs map[rtypes.CoinOutputID]rtypes.CoinOutput, confirmed bool) (et ExplorerTransaction) {
	// Get the header information for the transaction.
	et.ID = txn.ID()
	et.Height = height
	et.Parent = block.ParentID
	et.RawTransaction = txn
	et.Timestamp = block.Timestamp

	for k, tx := range block.Transactions {
		if et.ID == tx.ID() {
			et.Order = k
			break
		}
	}

	// go through the coin outputs to get the computation time for coin inputs
	var feeComputationTime rtypes.Timestamp
	for _, co := range txn.CoinOutputs {
		if co.Condition.ConditionType() == cftypes.ConditionTypeCustodyFee {
			feeComputationTime = co.Condition.Condition.(*cftypes.CustodyFeeCondition).ComputationTime
			break
		}
	}

	// Add the siacoin outputs that correspond with each siacoin input.
	for _, sci := range txn.CoinInputs {
		sco, ok := spentCoinOutputs[sci.ParentID]
		if !ok {
			build.Severe("could not find corresponding coin output")
		}
		feeInfo, err := getCoinOutputCustodyFeeInfo(plugin, feeComputationTime, sci.ParentID)
		if err != nil {
			build.Severe("error while fetching custody info for coin output", err)
		}
		et.CoinInputOutputs = append(et.CoinInputOutputs, ExplorerCoinOutput{
			ExplorerCoinOutput: rapi.ExplorerCoinOutput{
				CoinOutput: sco,
				UnlockHash: sco.Condition.UnlockHash(),
			},
			Custody: feeInfo,
		})
	}

	for i, co := range txn.CoinOutputs {
		coid := txn.CoinOutputID(uint64(i))
		et.CoinOutputIDs = append(et.CoinOutputIDs, coid)
		et.CoinOutputUnlockHashes = append(et.CoinOutputUnlockHashes, co.Condition.UnlockHash())
		if confirmed {
			feeInfo, err := getCoinOutputCustodyFeeInfo(plugin, chainTime, coid)
			if err != nil {
				build.Severe("error while fetching custody info for coin output", err)
			}
			et.CoinOutputCustodyFees = append(et.CoinOutputCustodyFees, feeInfo)
		}
	}

	// Add the siafund outputs that correspond to each siacoin input.
	for _, sci := range txn.BlockStakeInputs {
		sco, exists := explorer.BlockStakeOutput(sci.ParentID)
		if !exists {
			build.Severe("could not find corresponding blockstake output")
		}
		et.BlockStakeInputOutputs = append(et.BlockStakeInputOutputs, rapi.ExplorerBlockStakeOutput{
			BlockStakeOutput: sco,
			UnlockHash:       sco.Condition.UnlockHash(),
		})
	}

	for i, bso := range txn.BlockStakeOutputs {
		et.BlockStakeOutputIDs = append(et.BlockStakeOutputIDs, txn.BlockStakeOutputID(uint64(i)))
		et.BlockStakeOutputUnlockHashes = append(et.BlockStakeOutputUnlockHashes, bso.Condition.UnlockHash())
	}

	return et
}

func getCoinOutputCustodyFeeInfo(plugin *custodyfees.Plugin, chainTime rtypes.Timestamp, coid rtypes.CoinOutputID) (CustodyFeeInfo, error) {
	info, err := plugin.GetCoinOutputInfo(coid, chainTime)
	if err != nil {
		// acceptable in case coin output has already been spent
		return CustodyFeeInfo{}, nil
	}
	return CustodyFeeInfo{
		CreationTime:       info.CreationTime,
		CreationValue:      info.CreationValue,
		IsCustodyFee:       info.IsCustodyFee,
		Spent:              info.Spent,
		FeeComputationTime: info.FeeComputationTime,
		CustodyFee:         info.CustodyFee,
		SpendableValue:     info.SpendableValue,
	}, nil
}

type explorerTransactionsByHeight []ExplorerTransaction

func (h explorerTransactionsByHeight) Len() int      { return len(h) }
func (h explorerTransactionsByHeight) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h explorerTransactionsByHeight) Less(i, j int) bool {
	// Sort transactions in same block based of first appearance
	if h[i].Height == h[j].Height {
		return h[i].Order < h[j].Order
	}
	return h[i].Height < h[j].Height
}
