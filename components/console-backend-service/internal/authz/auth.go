package authz

import (
	"context"
	"reflect"
	"strings"
	"time"

	extractor "github.com/kyma-project/kyma/components/console-backend-service/internal/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/client-go/discovery"
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

type extractedAttributes struct {
	Name       string
	Namespace  string
	Resource   string
	APIGroup   string
	APIVersion string
}

// PrepareAttributes prepares attributes for authorization
func PrepareAttributes(ctx context.Context, u user.Info, attributes gqlschema.ResourceAttributes, obj interface{}, client discovery.DiscoveryInterface) (authorizer.Attributes, error) {
	resolverCtx := graphql.GetResolverContext(ctx)

	// make sure resource information is taken either from directive or from field
	err := validateAttributes(attributes)
	if err != nil {
		return nil, errors.Wrap(err, "while validating resource attributes")
	}

	var extracted extractedAttributes
	if attributes.IsChildResolver {
		extracted, err = extractAttributesFromChildResolver(obj, attributes)
	} else {
		extracted, err = extractAttributes(attributes, resolverCtx, client)
	}
	if err != nil {
		return nil, errors.Wrap(err, "while extracting attributes")
	}

	if attributes.ResourceArg == nil {
		extracted.Resource = *attributes.Resource
		extracted.APIVersion = *attributes.APIVersion
		extracted.APIGroup = *attributes.APIGroup
	}

	return authorizer.AttributesRecord{
		User:            u,
		Verb:            attributes.Verb,
		Namespace:       extracted.Namespace,
		APIGroup:        extracted.APIGroup,
		APIVersion:      extracted.APIVersion,
		Resource:        extracted.Resource,
		Subresource:     attributes.Subresource,
		Name:            extracted.Name,
		ResourceRequest: true,
	}, nil
}

func validateAttributes(attributes gqlschema.ResourceAttributes) error {
	if attributes.ResourceArg != nil {
		if attributes.Resource != nil || attributes.APIVersion != nil || attributes.APIGroup != nil {
			return errors.New("resource information shouldn't be both passed directly and extracted from field at the same time")
		}
		if attributes.IsChildResolver {
			return errors.New("resource information can't be extracted from passed object")
		}
	} else {
		if attributes.Resource == nil || attributes.APIVersion == nil || attributes.APIGroup == nil {
			return errors.New("resource information is missing, it should be either passed directly or extracted from field")
		}
	}

	return nil
}

func extractAttributesFromChildResolver(obj interface{}, attributes gqlschema.ResourceAttributes) (extractedAttributes, error) {
	var extracted extractedAttributes

	val := reflect.Indirect(reflect.ValueOf(obj))

	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name

		if attributes.NameArg != nil {
			if nameArg := *attributes.NameArg; fieldName == nameArg {
				// if field does not contain string
				if val.Field(i).Kind() != reflect.String {
					return extractedAttributes{}, errors.New("there are problems with receiving name value from passed object")
				}
				extracted.Name = val.Field(i).String()
			}
		}

		if attributes.NamespaceArg != nil {
			if namespaceArg := *attributes.NamespaceArg; fieldName == namespaceArg {
				// if field does not contain string
				if val.Field(i).Kind() != reflect.String {
					return extractedAttributes{}, errors.New("there are problems with receiving namespace value from passed object")
				}
				extracted.Namespace = val.Field(i).String()
			}
		}
	}

	if attributes.NameArg != nil && extracted.Name == "" {
		return extractedAttributes{}, errors.New("name field in passed object not found")
	}
	if attributes.NamespaceArg != nil && extracted.Namespace == "" {
		return extractedAttributes{}, errors.New("namespace field in passed object not found")
	}

	return extracted, nil
}

func extractAttributes(attributes gqlschema.ResourceAttributes, resolverCtx *graphql.ResolverContext, client discovery.DiscoveryInterface) (extractedAttributes, error) {
	var extracted extractedAttributes

	if attributes.NameArg != nil {
		name, err := extractValue(resolverCtx.Args[*attributes.NameArg], "name")
		if err != nil {
			return extractedAttributes{}, err
		}
		extracted.Name = name
	}

	if attributes.NamespaceArg != nil {
		namespace, err := extractValue(resolverCtx.Args[*attributes.NamespaceArg], "namespace")
		if err != nil {
			return extractedAttributes{}, err
		}
		extracted.Namespace = namespace
	}

	if attributes.ResourceArg != nil {
		resourceJSON, ok := resolverCtx.Args[*attributes.ResourceArg].(gqlschema.JSON)
		if !ok {
			return extractedAttributes{}, errors.New("resource in arguments found, but can't be converted to JSON")
		}
		resourceMeta, err := extractor.ExtractResourceMeta(resourceJSON)
		if err != nil {
			return extractedAttributes{}, errors.Wrap(err, "exile extracting meta information from resource meta")
		}
		split := strings.Split(resourceMeta.APIVersion, "/")
		switch len(split) {
		case 2:
			extracted.APIGroup = split[0]
			extracted.APIVersion = split[1]
		case 1:
			extracted.APIGroup = ""
			extracted.APIVersion = split[0]
		default:
			return extractedAttributes{}, errors.Errorf("resource apiVersion %q is invalid. Expected either one or no `/` delimiter", resourceMeta.APIVersion)
		}
		extracted.Resource, err = extractor.GetPluralNameFromKind(resourceMeta.Kind, resourceMeta.APIVersion, client)
		if err != nil {
			return extractedAttributes{}, errors.Wrap(err, "while getting resource's plural name")
		}
	}

	return extracted, nil
}

func extractValue(arg interface{}, valueName string) (string, error) {
	var value string
	switch v := arg.(type) {
	case string:
		value = v
	case *string:
		if v == nil {
			return "", nil
		}
		value = *v
	default:
		return "", errors.Errorf("%s in arguments found, but can't be converted to string", valueName)
	}
	if value == "" {
		return "", errors.Errorf("%s in arguments not found", valueName)
	}
	return value, nil
}
