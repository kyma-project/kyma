package tokens

import (
	"time"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/tokens/tokencache"
)

type TokenCreatorProvider interface {
	WithTTL(ttl time.Duration) Creator
}

type tokenCreatorProvider struct {
	store     tokencache.TokenCache
	generator Generator
}

func (tsp tokenCreatorProvider) WithTTL(ttl time.Duration) Creator {
	return NewTokenCreator(ttl, tsp.store, tsp.generator)
}

func NewTokenCreatorProvider(store tokencache.TokenCache, generator Generator) *tokenCreatorProvider {
	return &tokenCreatorProvider{
		store:     store,
		generator: generator,
	}
}
