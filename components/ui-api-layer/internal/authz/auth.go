package authz

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
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
	resolverCtx := graphql.GetResolverContext(ctx)

	var name string
	if attributes.NameArg != nil {
		name = resolverCtx.Args[*attributes.NameArg].(string)
	}

	var namespace string
	if attributes.NamespaceArg != nil {
		namespace = resolverCtx.Args[*attributes.NamespaceArg].(string)
	}

	return authorizer.AttributesRecord{
		User:            u,
		Verb:            attributes.Verb,
		Namespace:       namespace,
		APIGroup:        attributes.APIGroup,
		APIVersion:      attributes.APIVersion,
		Resource:        attributes.Resource,
		Subresource:     attributes.Subresource,
		Name:            name,
		ResourceRequest: true,
	}
}
