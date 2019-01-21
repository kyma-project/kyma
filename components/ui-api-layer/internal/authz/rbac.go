package authz

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/authn"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

type RBACDirective func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.ResourceAttributes) (res interface{}, err error)

func NewRBACDirective(a authorizer.Authorizer) RBACDirective {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.ResourceAttributes) (res interface{}, err error) {
		u := authn.UserInfoForContext(ctx)
		attrs := PrepareAttributes(ctx, u, attributes)
		authorized, _, err := a.Authorize(attrs)

		if authorized != authorizer.DecisionAllow {
			if err != nil {
				glog.Errorf("Error during authorization: %v", err)
			}
			return nil, errors.New("access denied")
		}

		return next(ctx)
	}
}
