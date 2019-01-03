package tokens

//type Cache interface {
//	Put(app string, tokenData *TokenData)
//	Get(app string) (*TokenData, bool)
//	Delete(app string)
//}
//
//type tokenCache struct {
//	tokenCache *cache.Cache
//}
//
//// TODO - decide if there should be separate cache for the Runtime tokens and Application tokens
//func NewTokenCache(expirationMinutes int) Cache {
//	return &tokenCache{
//		tokenCache: cache.New(time.Duration(expirationMinutes)*time.Minute, 1*time.Minute),
//	}
//}
//
//func (c *tokenCache) Put(app string, tokenData *TokenData) {
//	c.tokenCache.Set(app, tokenData, cache.DefaultExpiration)
//}
//
//func (c *tokenCache) Get(app string) (*TokenData, bool) {
//	token, found := c.tokenCache.Get(app)
//	if !found {
//		return nil, found
//	}
//
//	return token.(*TokenData), found
//}
//
//func (c *tokenCache) Delete(app string) {
//	c.tokenCache.Delete(app)
//}
