package authz

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	. "github.com/smartystreets/goconvey/convey"
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

func verifyCommonAttributes(authAttributes authorizer.Attributes) {
	Convey("Then user is set", func() {
		So(authAttributes.GetUser(), ShouldEqual, &userInfo)
	})
	Convey("Then verb is set", func() {
		So(authAttributes.GetVerb(), ShouldEqual, verb)
	})
	Convey("Then API group is set", func() {
		So(authAttributes.GetAPIGroup(), ShouldEqual, apiGroup)
	})
	Convey("Then API version is set", func() {
		So(authAttributes.GetAPIVersion(), ShouldEqual, apiVersion)
	})
	Convey("Then resource is set", func() {
		So(authAttributes.GetResource(), ShouldEqual, resource)
	})
	Convey("Then subresource is set", func() {
		So(authAttributes.GetSubresource(), ShouldEqual, subresource)
	})
	Convey("Then ResourceRequest is true", func() {
		So(authAttributes.IsResourceRequest(), ShouldBeTrue)
	})
}
