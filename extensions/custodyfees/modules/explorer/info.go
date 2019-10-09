package explorer

// LatestChainFacts returns the last known aggregated chain facts.
func (e *Explorer) LatestChainFacts() (facts ChainFacts, err error) {
	err = e.db.View(dbGetChainFactsDataFunc(&facts))
	return
}
