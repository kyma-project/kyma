package processor

func (e *Processor) PopulateErrors(errorsCh chan *ResultError) []ResultError {
	return e.populateErrors(errorsCh)
}

func (e *Processor) PopulateResults(resultsCh chan *ResultSuccess) []ResultSuccess {
	return e.populateResults(resultsCh)
}
