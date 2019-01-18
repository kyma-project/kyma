package authz

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	verb = "list"
	apiGroup = "k8s.io"
	apiVersion = "v1alpha1"
	resource = "pod"
	subresource = "logs"
	namespace = "x-system"
	name = "x-deployment"
)

var (
	namespaceArg = "my-namespace"
	nameArg = "my-name"
	userInfo = user.DefaultInfo{Name: "Test User", UID: "deadbeef", Groups: []string{"admins", "testers"}}
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
	Convey("User is set", func() {
		So(authAttributes.GetUser(), ShouldEqual, &userInfo)
	})
	Convey("Verb is set", func() {
		So(authAttributes.GetVerb(), ShouldEqual, verb)
	})
	Convey("API group is set", func() {
		So(authAttributes.GetAPIGroup(), ShouldEqual, apiGroup)
	})
	Convey("API version is set", func() {
		So(authAttributes.GetAPIVersion(), ShouldEqual, apiVersion)
	})
	Convey("Resource is set", func() {
		So(authAttributes.GetResource(), ShouldEqual, resource)
	})
	Convey("Subresource is set", func() {
		So(authAttributes.GetSubresource(), ShouldEqual, subresource)
	})
	Convey("ResourceRequest is true", func() {
		So(authAttributes.IsResourceRequest(), ShouldBeTrue)
	})
}