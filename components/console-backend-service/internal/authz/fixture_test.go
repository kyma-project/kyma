package authz

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

const (
	verb        = "list"
	subresource = "logs"
	namespace   = "x-system"
	name        = "x-deployment"
	kind        = "Deployment"
)

var (
	apiGroup     = "apps"
	apiVersion   = "v1"
	groupVersion = fmt.Sprintf("%s/%s", apiGroup, apiVersion)
	resource     = "deployments"
	resourceArg  = "MyResource"
	namespaceArg = "MyNamespace"
	nameArg      = "MyName"
	userInfo     = user.DefaultInfo{Name: "Test User", UID: "deadbeef", Groups: []string{"admins", "testers"}}
	resourceJSON = gqlschema.JSON{
		"apiVersion": groupVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	}
	noGroupResourceJSON = gqlschema.JSON{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	}
	fakeResources = []*v1.APIResourceList{
		&v1.APIResourceList{
			TypeMeta:     v1.TypeMeta{},
			GroupVersion: groupVersion,
			APIResources: []v1.APIResource{
				v1.APIResource{
					Name: resource,
					Kind: kind,
				},
			},
		},
		&v1.APIResourceList{
			TypeMeta:     v1.TypeMeta{},
			GroupVersion: apiVersion,
			APIResources: []v1.APIResource{
				v1.APIResource{
					Name: resource,
					Kind: kind,
				},
			},
		},
	}
)

type ChildResolverSetting bool

const (
	withChildResolverSet    = true
	withoutChildResolverSet = false
)

func noArgsAttributes(isChildResolver ChildResolverSetting) gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:            verb,
		APIGroup:        &apiGroup,
		APIVersion:      &apiVersion,
		Resource:        &resource,
		ResourceArg:     nil,
		Subresource:     subresource,
		NameArg:         nil,
		NamespaceArg:    nil,
		IsChildResolver: bool(isChildResolver),
	}
}

func withArgsAttributes(isChildResolver ChildResolverSetting) gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:            verb,
		APIGroup:        &apiGroup,
		APIVersion:      &apiVersion,
		Resource:        &resource,
		ResourceArg:     nil,
		Subresource:     subresource,
		NameArg:         &nameArg,
		NamespaceArg:    &namespaceArg,
		IsChildResolver: bool(isChildResolver),
	}
}

func withNamespaceArgAttributes(isChildResolver ChildResolverSetting) gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:            verb,
		APIGroup:        &apiGroup,
		APIVersion:      &apiVersion,
		Resource:        &resource,
		ResourceArg:     nil,
		Subresource:     subresource,
		NameArg:         nil,
		NamespaceArg:    &namespaceArg,
		IsChildResolver: bool(isChildResolver),
	}
}

func withResourceArgAttributes(isChildResolver ChildResolverSetting) gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:            verb,
		APIGroup:        nil,
		APIVersion:      nil,
		Resource:        nil,
		ResourceArg:     &resourceArg,
		Subresource:     subresource,
		NameArg:         nil,
		NamespaceArg:    nil,
		IsChildResolver: bool(isChildResolver),
	}
}

func noResourceAttributes(isChildResolver ChildResolverSetting) gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:            verb,
		APIGroup:        nil,
		APIVersion:      nil,
		Resource:        nil,
		ResourceArg:     nil,
		Subresource:     subresource,
		NameArg:         nil,
		NamespaceArg:    nil,
		IsChildResolver: bool(isChildResolver),
	}
}

func withRedundantResourceArgAttributes(isChildResolver ChildResolverSetting) gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:            verb,
		APIGroup:        &apiGroup,
		APIVersion:      &apiVersion,
		Resource:        &resource,
		ResourceArg:     &resourceArg,
		Subresource:     subresource,
		NameArg:         nil,
		NamespaceArg:    nil,
		IsChildResolver: bool(isChildResolver),
	}
}

func noArgsContext() context.Context {
	resolver := graphql.ResolverContext{Args: map[string]interface{}{}}
	return graphql.WithResolverContext(context.Background(), &resolver)
}

func withArgsContext(resource gqlschema.JSON) context.Context {
	resolver := graphql.ResolverContext{Args: map[string]interface{}{
		namespaceArg: namespace,
		nameArg:      name,
		resourceArg:  resource,
	}}
	return graphql.WithResolverContext(context.Background(), &resolver)
}

func verifyCommonAttributes(t *testing.T, authAttributes authorizer.Attributes) {
	require.NotNil(t, authAttributes)
	t.Run("Then user is set", func(t *testing.T) {
		assert.Equal(t, &userInfo, authAttributes.GetUser())
	})
	t.Run("Then verb is set", func(t *testing.T) {
		assert.Equal(t, verb, authAttributes.GetVerb())
	})
	t.Run("Then API group is set", func(t *testing.T) {
		assert.Equal(t, apiGroup, authAttributes.GetAPIGroup())
	})
	t.Run("Then API version is set", func(t *testing.T) {
		assert.Equal(t, apiVersion, authAttributes.GetAPIVersion())
	})
	t.Run("Then resource is set", func(t *testing.T) {
		assert.Equal(t, resource, authAttributes.GetResource())
	})
	t.Run("Then subresource is set", func(t *testing.T) {
		assert.Equal(t, subresource, authAttributes.GetSubresource())
	})
	t.Run("Then ResourceRequest is true", func(t *testing.T) {
		assert.True(t, authAttributes.IsResourceRequest())
	})
}
