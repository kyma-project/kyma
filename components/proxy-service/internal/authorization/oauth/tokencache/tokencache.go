package tokencache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

type TokenCache interface {
	Get(clientID string) (token string, found bool)
	Add(clientID, token string, expirationSeconds int)
	Remove(clientID string)
}

type tokenCache struct {
	cache *cache.Cache
}

func NewTokenCache() TokenCache {
	return &tokenCache{
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
}

func (tc *tokenCache) Get(clientId string) (token string, found bool) {
	res, found := tc.cache.Get(clientId)
	if !found {
		return "", false
	}

	return res.(string), found
}

func (tc *tokenCache) Add(clientID, token string, expirationSeconds int) {
	tc.cache.Set(clientID, token, time.Duration(expirationSeconds-2)*time.Second)
}

func (tc *tokenCache) Remove(clientID string) {
	tc.cache.Delete(clientID)
}
