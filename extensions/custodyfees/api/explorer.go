package api

import (
	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
)

// RegisterExplorerCustodyFeesHTTPHandlers registers the default explorer HTTP handlers specific to the custodyfees package.
func RegisterExplorerCustodyFeesHTTPHandlers(router rapi.Router, cs modules.ConsensusSet, plugin *custodyfees.Plugin) {
	router.GET("/explorer/custodyfees/coinoutput/:id", NewCoinOutputInfoGetHandler(cs, plugin))
}
