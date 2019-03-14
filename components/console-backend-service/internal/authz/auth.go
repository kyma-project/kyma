package authz

import (
	"context"
	"reflect"
	"strings"
	"time"

	extractor "github.com/kyma-project/kyma/components/console-backend-service/internal/extractor"
	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"

	"github.com/99designs/gqlgen/graphql"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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
func PrepareAttributes(ctx context.Context, u user.Info, attributes gqlschema.ResourceAttributes, obj interface{}, client discovery.DiscoveryInterface) (authorizer.Attributes, error) {
	resolverCtx := graphql.GetResolverContext(ctx)

	var name string
	var namespace string
	var resource string
	var apiGroup string
	var apiVersion string

	// make sure resource information is taken either from directive or from field
	if attributes.ResourceArg != nil {
		if attributes.Resource != nil || attributes.APIVersion != nil || attributes.APIGroup != nil {
			return nil, errors.New("resource information shouldn't be both passed directly and extracted from field at the same time")
		}
		if attributes.IsChildResolver {
			return nil, errors.New("resource information can't be extracted from passed object")
		}
	} else {
		if attributes.Resource == nil || attributes.APIVersion == nil || attributes.APIGroup == nil {
			return nil, errors.New("resource information is missing, it should be either passed directly or extracted from field")
		}
		resource = *attributes.Resource
		apiGroup = *attributes.APIGroup
		apiVersion = *attributes.APIVersion
	}

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

		if attributes.ResourceArg != nil {
			resourceJSON, ok := resolverCtx.Args[*attributes.ResourceArg].(gqlschema.JSON)
			if !ok {
				return nil, errors.New("resource in arguments found, but can't be converted to JSON")
			}
			resourceMeta, err := extractor.ExtractResourceMeta(resourceJSON)
			if err != nil {
				return nil, errors.New("resource in arguments found, but meta information can't be extracted")
			}
			split := strings.Split(resourceMeta.APIVersion, "/")
			switch len(split) {
			case 2:
				apiGroup = split[0]
				apiVersion = split[1]
			case 1:
				apiGroup = ""
				apiVersion = split[0]
			default:
				return nil, errors.New("resource apiVersion format is invalid")
			}
			resource, err = extractor.GetPluralNameFromKind(resourceMeta.Kind, resourceMeta.APIVersion, client)
			if err != nil {
				return nil, errors.Wrap(err, "while getting resource's plural name")
			}
		}
	}

	return authorizer.AttributesRecord{
		User:            u,
		Verb:            attributes.Verb,
		Namespace:       namespace,
		APIGroup:        apiGroup,
		APIVersion:      apiVersion,
		Resource:        resource,
		Subresource:     attributes.Subresource,
		Name:            name,
		ResourceRequest: true,
	}, nil
}
