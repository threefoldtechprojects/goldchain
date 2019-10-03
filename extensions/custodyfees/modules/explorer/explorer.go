// Package explorer provides a glimpse into what the network currently
// looks like.
package explorer

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"
)

var (
	errNilCS = errors.New("explorer cannot use a nil consensus set")
)

type (
	// Explorer is a Custody Fee Explorer Keeps some global metrics about custody fees.
	Explorer struct {
		cs         modules.ConsensusSet
		db         *persist.BoltDatabase
		log        *persist.Logger
		persistDir string
		bcInfo     types.BlockchainInfo
		chainCts   types.ChainConstants
	}
)

// New creates the internal data structures, and subscribes to
// consensus for changes to the blockchain
func New(cs modules.ConsensusSet, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verboseLogging bool) (*Explorer, error) {
	// Check that input modules are non-nil
	if cs == nil {
		return nil, errNilCS
	}

	// Initialize the explorer.
	e := &Explorer{
		cs:         cs,
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
