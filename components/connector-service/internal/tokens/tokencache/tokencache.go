package tokencache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const defaultTTLMinutes = 5

type TokenCache interface {
	Put(token string, data string, ttl time.Duration)
	Get(token string) (string, bool)
	Delete(token string)
}

type tokenCache struct {
	tokenCache *cache.Cache
}

func NewTokenCache() TokenCache {
	return &tokenCache{
		tokenCache: cache.New(time.Duration(defaultTTLMinutes)*time.Minute, 1*time.Minute),
	}
}

func (c *tokenCache) Put(token string, data string, ttl time.Duration) {
	c.tokenCache.Set(token, data, ttl)
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
