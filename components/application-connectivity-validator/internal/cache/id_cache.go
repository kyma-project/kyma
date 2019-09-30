package cache

import (
	"time"

	cache2 "github.com/patrickmn/go-cache"
)

//go:generate mockery -name=IdCache
type IdCache interface {
	GetClientIDs(applicationName string) ([]string, bool)
	SetClientIDs(applicationName string, clientIDs []string)
}

type cache struct {
	cache *cache2.Cache
}

func NewCache(expirationMinutes, cleanupMinutes int) IdCache {
	expirationTime := time.Duration(expirationMinutes) * time.Minute
	cleanupTime := time.Duration(cleanupMinutes) * time.Minute

	return &cache{
		cache: cache2.New(expirationTime, cleanupTime),
	}
}

func (c *cache) GetClientIDs(applicationName string) ([]string, bool) {
	clientIDs, found := c.cache.Get(applicationName)
	if !found {
		return []string{}, found
	}
	return clientIDs.([]string), found
}

func (c *cache) SetClientIDs(applicationName string, clientIDs []string) {
	c.cache.Set(applicationName, clientIDs, cache2.DefaultExpiration)
}
