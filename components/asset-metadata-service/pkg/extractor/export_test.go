package extractor

import "github.com/kyma-project/kyma/components/asset-metadata-service/pkg/matador"

func (e *Extractor) SetMatador(matador matador.Matador) {
	e.matador = matador
}

func (e *Extractor) PopulateErrors(errorsCh chan *ResultError) []ResultError {
	return e.populateErrors(errorsCh)
}

func (e *Extractor) PopulateResults(resultsCh chan *ResultSuccess) []ResultSuccess {
	return e.populateResults(resultsCh)
}
