package jwt

import (
	"context"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/model"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
)

const (
	alg        = jwa.RS256
	ctxTimeout = time.Second * 10
)

const (
	claimEmail     = "email"
	claimScope     = "scope"
	scopeDeveloper = "runtimeDeveloper"
	scopeAdmin     = "runtimeNamespaceAdmin"
)

//go:generate mockery --name=JWX --output=automock --outpkg=automock --case=underscore
type JWX interface {
	Fetch(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error)
	Parse(s []byte, options ...jwt.ParseOption) (jwt.Token, error)
}

type Service struct {
	jwx JWX
}

func NewService() Service {
	return Service{
		jwx: defaultJWX{},
	}
}

func (s Service) ParseAndVerify(jwtSrc, jwksURI string) (jwt.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	keyset, err := s.jwx.Fetch(ctx, jwksURI)
	if err != nil {
		return nil, err
	}

	key, ok := keyset.Get(0)
	if !ok {
		return nil, errors.New("JWK key not found in key set")
	}

	parsed, err := s.jwx.Parse([]byte(jwtSrc), jwt.WithValidate(true), jwt.WithVerify(alg, key))
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func (s Service) GetUser(token jwt.Token) (model.User, error) {
	user := model.User{}

	email, ok := token.PrivateClaims()[claimEmail].(string)
	if !ok {
		return model.User{}, errors.New("Token does not contain valid email private claim")
	}
	user.Email = email

	scopes, ok := token.PrivateClaims()[claimScope].([]interface{})
	if !ok {
		return model.User{}, errors.New("Token does not contain valid scope private claim")
	}

	for _, v := range scopes {
		scope, ok := v.(string)
		if !ok {
			continue
		}

		if strings.Contains(scope, scopeDeveloper) {
			user.IsDeveloper = true
		} else if strings.Contains(scope, scopeAdmin) {
			user.IsAdmin = true
		}
	}

	return user, nil
}
