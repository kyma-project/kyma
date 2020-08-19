package authz

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/client-go/discovery"
)

type RBACDirective func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.ResourceAttributes) (res interface{}, err error)

func NewRBACDirective(a authorizer.Authorizer, client discovery.DiscoveryInterface) RBACDirective {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.ResourceAttributes) (res interface{}, err error) {
		u, err := authn.UserInfoForContext(ctx)
		if err != nil {
			glog.Errorf("Error while receiving user information: %v", err)
			return nil, errors.New("access denied due to problems on the server side")
		}

		attrs, err := PrepareAttributes(ctx, u, attributes, obj, client)
		if err != nil {
			glog.Errorf("Error while obtaining attributes for authorization: %v", err)
			return nil, errors.New("access denied due to problems on the server side")
		}
		authorized, _, err := a.Authorize(ctx, attrs)

		if authorized != authorizer.DecisionAllow {
			if err != nil {
				glog.Errorf("Error during authorization: %v", err)
			}
			return nil, errors.New("access denied")
		}

		return next(ctx)
	}
}
