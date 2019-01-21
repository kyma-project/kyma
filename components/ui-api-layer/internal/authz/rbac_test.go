package authz

import (
	"context"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/authn"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRBACDirective(t *testing.T) {
	ctx := authn.WithUserInfoContext(withArgsContext(), &userInfo)

	Convey("When decision is 'allow'", t, func() {
		authorizerMock := &mockAuthorizer{decision: authorizer.DecisionAllow}
		resolver := &mockResolver{}

		attributes := withArgsAttributes()

		directive := NewRBACDirective(authorizerMock)

		_, err := directive(ctx, 0, resolver.Mock, attributes)

		Convey("Then authorizer is called with proper attributes", func() {
			verifyCommonAttributes(authorizerMock.lastCallAttributes)
		})
		Convey("Then next resolver is called with input context", func() {
			So(resolver.lastCallContext, ShouldEqual, ctx)
		})
		Convey("Then no error is returned", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When decision is 'deny'", t, func() {
		authorizerMock := &mockAuthorizer{decision: authorizer.DecisionDeny}
		resolver := &mockResolver{}
		attributes := withArgsAttributes()

		directive := NewRBACDirective(authorizerMock)

		_, err := directive(ctx, 0, resolver.Mock, attributes)

		Convey("Then authorizer is called with proper attributes", func() {
			verifyCommonAttributes(authorizerMock.lastCallAttributes)
		})
		Convey("Then next resolver is not called", func() {
			So(resolver.lastCallContext, ShouldEqual, nil)
		})
		Convey("Then error is returned", func() {
			So(err, ShouldNotBeNil)
		})
	})

	Convey("When decision is 'no opinion'", t, func() {
		authorizerMock := &mockAuthorizer{decision: authorizer.DecisionNoOpinion}
		resolver := &mockResolver{}
		attributes := withArgsAttributes()

		directive := NewRBACDirective(authorizerMock)

		_, err := directive(ctx, 0, resolver.Mock, attributes)

		Convey("Then authorizer is called with proper attributes", func() {
			verifyCommonAttributes(authorizerMock.lastCallAttributes)
		})
		Convey("Then next resolver is not called", func() {
			So(resolver.lastCallContext, ShouldEqual, nil)
		})
		Convey("Then error is returned", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

type mockAuthorizer struct {
	authorizer.Authorizer

	decision           authorizer.Decision
	lastCallAttributes authorizer.Attributes
}

func (a *mockAuthorizer) Authorize(attrs authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	a.lastCallAttributes = attrs
	return a.decision, "", nil
}

type mockResolver struct {
	lastCallContext context.Context
}

func (r *mockResolver) Mock(ctx context.Context) (interface{}, error) {
	r.lastCallContext = ctx
	return 0, nil
}
