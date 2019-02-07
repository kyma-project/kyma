package authz

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/99designs/gqlgen/graphql"

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

type SARCacheConfig struct {
	AllowCacheTTL time.Duration `envconfig:"default=5m"`
	DenyCacheTTL  time.Duration `envconfig:"default=30s"`
}

func NewAuthorizer(client authorizationclient.SubjectAccessReviewInterface, cacheConfig SARCacheConfig) (authorizer.Authorizer, error) {
	if client == nil {
		return nil, errors.New("no client provided, cannot use webhook authorization")
	}
	authorizerConfig := authorizerfactory.DelegatingAuthorizerConfig{
		SubjectAccessReviewClient: client,
		AllowCacheTTL:             cacheConfig.AllowCacheTTL,
		DenyCacheTTL:              cacheConfig.DenyCacheTTL,
	}
	a, err := authorizerConfig.New()

	return &Authorizer{a}, err
}

// PrepareAttributes prepares attributes for authorization
func PrepareAttributes(ctx context.Context, u user.Info, attributes gqlschema.ResourceAttributes, obj interface{}) (authorizer.Attributes, error) {
	resolverCtx := graphql.GetResolverContext(ctx)

	var name string
	var namespace string

	if attributes.IsChildResolver {
		val := reflect.Indirect(reflect.ValueOf(obj))

		for i := 0; i < val.NumField(); i++ {
			fieldName := val.Type().Field(i).Name

			if attributes.NameArg != nil {
				if nameArg := *attributes.NameArg; fieldName == nameArg {
					// if field does not contain string
					if val.Field(i).Kind() != reflect.String {
						return nil, errors.New("there are problems with receiving name value from passed object")
					}
					name = val.Field(i).String()
				}
			}

			if attributes.NamespaceArg != nil {
				if namespaceArg := *attributes.NamespaceArg; fieldName == namespaceArg {
					// if field does not contain string
					if val.Field(i).Kind() != reflect.String {
						return nil, errors.New("there are problems with receiving namespace value from passed object")
					}
					namespace = val.Field(i).String()
				}
			}
		}

		if attributes.NameArg != nil && name == "" {
			return nil, errors.New("name field in passed object not found")
		}
		if attributes.NamespaceArg != nil && namespace == "" {
			return nil, errors.New("namespace field in passed object not found")
		}
	} else {

		if attributes.NameArg != nil {
			var ok bool
			name, ok = resolverCtx.Args[*attributes.NameArg].(string)
			if !ok {
				return nil, errors.New("name in arguments found, but can't be converted to string")
			}
			if name == "" {
				return nil, errors.New("name in arguments not found")
			}
		}

		if attributes.NamespaceArg != nil {
			var ok bool
			namespace, ok = resolverCtx.Args[*attributes.NamespaceArg].(string)
			if !ok {
				return nil, errors.New("namespace in arguments found, but can't be converted to string")
			}
			if namespace == "" {
				return nil, errors.New("namespace in arguments not found")
			}
		}
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
	}, nil
}
