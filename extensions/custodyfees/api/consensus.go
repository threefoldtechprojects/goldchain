package api

import (
	"github.com/nbh-digital/goldchain/extensions/custodyfees"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
)

// RegisterConsensusCustodyFeesHTTPHandlers registers the default consensus HTTP handlers specific to the custodyfees package.
func RegisterConsensusCustodyFeesHTTPHandlers(router rapi.Router, cs modules.ConsensusSet, plugin *custodyfees.Plugin) {
	router.GET("/consensus/custodyfees/coinoutput/:id", NewCoinOutputInfoGetHandler(cs, plugin))
}
