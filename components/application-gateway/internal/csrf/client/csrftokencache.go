package client

import (
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"

	cache "github.com/patrickmn/go-cache"
)

//Cache for CSRF data items
type TokenCache interface {
	Get(itemID string) (resp *csrf.Response, found bool)
	Add(itemID string, resp *csrf.Response)
	Remove(itemID string)
}

//Creates a new TokenCache instance
func NewTokenCache() TokenCache {
	return &tokenCache{
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
}

type tokenCache struct {
	cache *cache.Cache
}

func (tc *tokenCache) Get(itemID string) (resp *csrf.Response, found bool) {
	res, found := tc.cache.Get(itemID)
	if !found {
		return nil, false
	}

	return res.(*csrf.Response), found
}

func (tc *tokenCache) Add(itemID string, resp *csrf.Response) {
	tc.cache.Set(itemID, resp, time.Duration(-1)*time.Second)
}

func (tc *tokenCache) Remove(itemID string) {
	tc.cache.Delete(itemID)
}
