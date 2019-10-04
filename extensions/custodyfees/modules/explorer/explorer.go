// Package explorer provides a glimpse into what the network currently
// looks like.
package explorer

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"

	"github.com/nbh-digital/goldchain/extensions/custodyfees"
)

type (
	// Explorer is a Custody Fee Explorer Keeps some global metrics about custody fees.
	Explorer struct {
		cs         modules.ConsensusSet
		plugin     *custodyfees.Plugin
		db         *persist.BoltDatabase
		log        *persist.Logger
		persistDir string
		bcInfo     types.BlockchainInfo
		chainCts   types.ChainConstants
	}
)

// New creates the internal data structures, and subscribes to
// consensus for changes to the blockchain
func New(cs modules.ConsensusSet, plugin *custodyfees.Plugin, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verboseLogging bool) (*Explorer, error) {
	if cs == nil {
		return nil, errors.New("no ConsensusSet given while one is required")
	}
	if plugin == nil {
		return nil, errors.New("no Custody Fees plugin given while one is required")
	}

	// Initialize the explorer.
	e := &Explorer{
		cs:         cs,
		plugin:     plugin,
		persistDir: persistDir,
		bcInfo:     bcInfo,
		chainCts:   chainCts,
	}

	// Initialize the persistent structures, including the database.
	err := e.initPersist(verboseLogging)
	if err != nil {
		return nil, err
	}

	// retrieve the current ConsensusChangeID
	var recentChange modules.ConsensusChangeID
	err = e.db.View(dbGetInternal(internalRecentChange, &recentChange))
	if err != nil {
		return nil, err
	}

	err = cs.ConsensusSetSubscribe(e, recentChange, nil)
	if err != nil {
		// TODO: restart from 0
		return nil, errors.New("explorer subscription failed: " + err.Error())
	}

	return e, nil
}

// Close closes the explorer.
func (e *Explorer) Close() error {
	e.cs.Unsubscribe(e)
	// Set up closing the logger.
	if e.log != nil {
		err := e.log.Close()
		if err != nil {
			// State of the logger is unknown, a println will suffice.
			fmt.Println("Error shutting down explorer logger:", err)
		}
	}
	return e.db.Close()
}
