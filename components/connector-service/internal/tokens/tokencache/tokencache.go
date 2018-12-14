package tokencache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type TokenCache interface {
	Put(app string, token string)
	Get(app string) (string, bool)
	Delete(app string)
}

type tokenCache struct {
	tokenCache *cache.Cache
}

func NewTokenCache(expirationMinutes int) TokenCache {
	return &tokenCache{
		tokenCache: cache.New(time.Duration(expirationMinutes)*time.Minute, 1*time.Minute),
	}
}

func (c *tokenCache) Put(app string, token string) {
	c.tokenCache.Set(app, token, cache.DefaultExpiration)
}

func (c *tokenCache) Get(app string) (string, bool) {
	token, found := c.tokenCache.Get(app)
	if !found {
		return "", found
	}

	return token.(string), found
}

func (c *tokenCache) Delete(app string) {
	c.tokenCache.Delete(app)
}
