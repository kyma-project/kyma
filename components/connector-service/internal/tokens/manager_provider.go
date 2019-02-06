package tokens

import (
	"time"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type TokenManagerProvider interface {
	WithTTL(ttl time.Duration) Manager
}

type tokenManagerProvider struct {
	store     tokencache.TokenCache
	generator Generator
}

func (tsp tokenManagerProvider) WithTTL(ttl time.Duration) Manager {
	return NewTokenManager(ttl, tsp.store, tsp.generator)
}

func NewTokenManagerProvider(store tokencache.TokenCache, generator Generator) *tokenManagerProvider {
	return &tokenManagerProvider{
		store:     store,
		generator: generator,
	}
}
