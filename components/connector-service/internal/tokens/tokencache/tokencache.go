package tokencache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type TokenCache interface {
	Put(re string, token string)
	Get(re string) (string, bool)
	Delete(re string)
}

type tokenCache struct {
	tokenCache *cache.Cache
}

func NewTokenCache(expirationMinutes int) TokenCache {
	return &tokenCache{
		tokenCache: cache.New(time.Duration(expirationMinutes)*time.Minute, 1*time.Minute),
	}
}

func (c *tokenCache) Put(re string, token string) {
	c.tokenCache.Set(re, token, cache.DefaultExpiration)
}

func (c *tokenCache) Get(re string) (string, bool) {
	token, found := c.tokenCache.Get(re)
	if !found {
		return "", found
	}

	return token.(string), found
}

func (c *tokenCache) Delete(re string) {
	c.tokenCache.Delete(re)
}
