package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

const (
	verb        = "list"
	apiGroup    = "k8s.io"
	apiVersion  = "v1alpha1"
	resource    = "pod"
	subresource = "logs"
	namespace   = "x-system"
	name        = "x-deployment"
)

var (
	namespaceArg = "my-namespace"
	nameArg      = "my-name"
	userInfo     = user.DefaultInfo{Name: "Test User", UID: "deadbeef", Groups: []string{"admins", "testers"}}
)

func noArgsAttributes() gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:         verb,
		APIGroup:     apiGroup,
		APIVersion:   apiVersion,
		Resource:     resource,
		Subresource:  subresource,
		NameArg:      nil,
		NamespaceArg: nil,
	}
}

func withArgsAttributes() gqlschema.ResourceAttributes {
	return gqlschema.ResourceAttributes{
		Verb:         verb,
		APIGroup:     apiGroup,
		APIVersion:   apiVersion,
		Resource:     resource,
		Subresource:  subresource,
		NameArg:      &nameArg,
		NamespaceArg: &namespaceArg,
	}
}

func noArgsContext() context.Context {
	resolver := graphql.ResolverContext{Args: map[string]interface{}{}}
	return graphql.WithResolverContext(context.Background(), &resolver)
}

func withArgsContext() context.Context {
	resolver := graphql.ResolverContext{Args: map[string]interface{}{
		namespaceArg: namespace,
		nameArg:      name,
	}}
	return graphql.WithResolverContext(context.Background(), &resolver)
}

func verifyCommonAttributes(t *testing.T, authAttributes authorizer.Attributes) {
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
