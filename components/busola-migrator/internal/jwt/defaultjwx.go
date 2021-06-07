package jwt

import (
	"context"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

type defaultJWX struct{}

func (j defaultJWX) Fetch(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
	return jwk.Fetch(ctx, urlstring, options...)
}

func (j defaultJWX) Parse(s []byte, options ...jwt.ParseOption) (jwt.Token, error) {
	return jwt.Parse(s, options...)
}
