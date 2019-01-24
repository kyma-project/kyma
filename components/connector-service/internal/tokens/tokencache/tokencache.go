package tokencache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type TokenCache interface {
	Put(token string, data string)
	Get(token string) (string, bool)
	Delete(token string)
}

type tokenCache struct {
	tokenCache *cache.Cache
}

func NewTokenCache(expirationMinutes int) TokenCache {
	return &tokenCache{
		tokenCache: cache.New(time.Duration(expirationMinutes)*time.Minute, 1*time.Minute),
	}
}

func (c *tokenCache) Put(token string, data string) {
	c.tokenCache.Set(token, data, cache.DefaultExpiration)
}

func (c *tokenCache) Get(token string) (string, bool) {
	data, found := c.tokenCache.Get(token)
	if !found {
		return "", found
	}

	return data.(string), found
}

func (c *tokenCache) Delete(token string) {
	c.tokenCache.Delete(token)
}
