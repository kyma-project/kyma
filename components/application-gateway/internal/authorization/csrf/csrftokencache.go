package csrf

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

type TokenCache interface {
	Get(itemID string) (resp *Response, found bool)
	Add(itemID string, resp *Response)
	Remove(itemID string)
}

type tokenCache struct {
	cache *cache.Cache
}

func NewTokenCache() TokenCache {
	return &tokenCache{
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
}

func (tc *tokenCache) Get(itemID string) (resp *Response, found bool) {
	res, found := tc.cache.Get(itemID)
	if !found {
		return nil, false
	}

	return res.(*Response), found
}

func (tc *tokenCache) Add(itemID string, resp *Response) {
	tc.cache.Set(itemID, resp, time.Duration(-1)*time.Second)
}

func (tc *tokenCache) Remove(itemID string) {
	tc.cache.Delete(itemID)
}
