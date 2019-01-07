package authz

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/authn"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	authorizerpkg "k8s.io/apiserver/pkg/authorization/authorizer"
)

type RBACDirective func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.RBACAttributes) (res interface{}, err error)

func NewRBACDirective(authorizer authorizerpkg.Authorizer) RBACDirective {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.RBACAttributes) (res interface{}, err error) {

		// fetch user from context
		u := authn.UserInfoForContext(ctx)

		// prepare attributes for authz
		attrs := PrepareAttributes(ctx, u, attributes)
		glog.Infof("SAR attributes: %+v", attrs)

		// check if user is allowed to get requested resource
		authorized, reason, err := authorizer.Authorize(attrs)
		glog.Infof("authorized: %v, reason: %s, err: %v", authorized, reason, err)

		if authorized != authorizerpkg.DecisionAllow {
			if err != nil {
				glog.Errorf("Error during authorization: %v", err)
			}
			return nil, errors.New("access denied")
		}

		return next(ctx)
	}
}
