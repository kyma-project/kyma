package authz

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
)

type Authorizer struct {

	// authorizer determines whether a given authorization.Attributes is allowed
	authorizer.Authorizer
}

func NewAuthorizer(client authorizationclient.SubjectAccessReviewInterface) (authorizer.Authorizer, error) {
	if client == nil {
		return nil, errors.New("no client provided, cannot use webhook authorization")
	}
	authorizerConfig := authorizerfactory.DelegatingAuthorizerConfig{
		SubjectAccessReviewClient: client,
		AllowCacheTTL:             5 * time.Minute,
		DenyCacheTTL:              30 * time.Second,
	}
	a, err := authorizerConfig.New()

	return &Authorizer{a}, err
}

// PrepareAttributes prepares attributes for authorization
func PrepareAttributes(ctx context.Context, u user.Info, attributes gqlschema.RBACAttributes) authorizer.Attributes {
	// TODO: implement
	return authorizer.AttributesRecord{}
}
