package processor

func (p *Processor) PopulateErrors(errorsCh chan *ResultError) []ResultError {
	return p.populateErrors(errorsCh)
}

func (p *Processor) PopulateResults(resultsCh chan *ResultSuccess) []ResultSuccess {
	return p.populateResults(resultsCh)
}
