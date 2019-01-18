package authz

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"testing"
)


func TestPrepareAttributes(t *testing.T) {

	verb := "list"
	apiGroup := "k8s.io"
	apiVersion := "v1alpha1"
	resource := "pod"
	subresource := "logs"
	namespaceArg := "my-namespace"
	namespace := "x-system"
	nameArg := "my-name"
	name := "x-deployment"
	userInfo := user.DefaultInfo{Name: "Test User", UID: "deadbeef", Groups: []string{"admins", "testers"}}

	defaultAttributes := func() gqlschema.ResourceAttributes {
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

	verifyDefault := func(authAttributes authorizer.Attributes) {
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

	ctx := context.Background()

	Convey("When no arg is required", t, func() {

		gqlAttributes := defaultAttributes()
		authAttributes := PrepareAttributes(ctx, &userInfo, gqlAttributes)

		verifyDefault(authAttributes)

		Convey("Namespace is empty", func() {
			So(authAttributes.GetNamespace(), ShouldBeEmpty)
		})
		Convey("Name is empty", func() {
			So(authAttributes.GetName(), ShouldBeEmpty)
		})

	})

	Convey("When args are required", t, func() {
		gqlAttributes := defaultAttributes()
		gqlAttributes.NamespaceArg = &namespaceArg
		gqlAttributes.NameArg = &nameArg
		resolver := graphql.ResolverContext{Args: map[string]interface{}{
			namespaceArg: namespace,
			nameArg: name,

		}}
		resolverCtx := graphql.WithResolverContext(ctx, &resolver)
		authAttributes := PrepareAttributes(resolverCtx, &userInfo, gqlAttributes)

		verifyDefault(authAttributes)
		Convey("Namespace is set", func() {
			So(authAttributes.GetNamespace(), ShouldEqual, namespace)
		})
		Convey("Name is set", func() {
			So(authAttributes.GetName(), ShouldEqual, name)
		})
	})
}