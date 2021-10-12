package httpconsts

const (
	HeaderContentType          = "Content-Type"
	HeaderAuthorization        = "Authorization"
	HeaderAccessToken          = "Access-Token"
	HeaderCSRFToken            = "X-csrf-token"
	HeaderUserAgent            = "User-Agent"
	HeaderXForwardedProto      = "X-Forwarded-Proto"
	HeaderXForwardedFor        = "X-Forwarded-For"
	HeaderXForwardedHost       = "X-Forwarded-Host"
	HeaderXForwardedClientCert = "X-Forwarded-Client-Cert"
	HeaderCSRFTokenVal         = "fetch"
	HeaderAccept               = "Accept"
	HeaderAcceptVal            = "*/*"
	HeaderCacheControl         = "cache-control"
	HeaderCacheControlVal      = "no-cache"
	HeaderCookie               = "Cookie"
)

const (
	ContentTypeApplicationJson       = "application/json;charset=UTF-8"
	ContentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"
)
