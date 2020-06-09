package authz

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

func TestNewRBACDirective(t *testing.T) {
	ctx := authn.WithUserInfoContext(withArgsContext(resourceJSON), &userInfo)

	t.Run("When decision is 'allow'", func(t *testing.T) {
		authorizerMock := &mockAuthorizer{decision: authorizer.DecisionAllow}
		resolver := &mockResolver{}

		attributes := withArgsAttributes(withoutChildResolverSet)

		clientset := fake.NewSimpleClientset()
		directive := NewRBACDirective(authorizerMock, clientset.Discovery())

		_, err := directive(ctx, 0, resolver.Mock, attributes)

		t.Run("Then authorizer is called with proper attributes", func(t *testing.T) {
			verifyCommonAttributes(t, authorizerMock.lastCallAttributes)
		})
		t.Run("Then next resolver is called with input context", func(t *testing.T) {
			assert.Equal(t, ctx, resolver.lastCallContext)
		})
		t.Run("Then no error is returned", func(t *testing.T) {
			assert.Nil(t, err)
		})
	})

	t.Run("When decision is 'deny'", func(t *testing.T) {
		authorizerMock := &mockAuthorizer{decision: authorizer.DecisionDeny}
		resolver := &mockResolver{}
		attributes := withArgsAttributes(withoutChildResolverSet)

		clientset := fake.NewSimpleClientset()
		directive := NewRBACDirective(authorizerMock, clientset.Discovery())

		_, err := directive(ctx, 0, resolver.Mock, attributes)

		t.Run("Then authorizer is called with proper attributes", func(t *testing.T) {
			verifyCommonAttributes(t, authorizerMock.lastCallAttributes)
		})
		t.Run("Then next resolver is not called", func(t *testing.T) {
			assert.Equal(t, nil, resolver.lastCallContext)
		})
		t.Run("Then error is returned", func(t *testing.T) {
			assert.NotNil(t, err)
		})
	})

	t.Run("When decision is 'no opinion'", func(t *testing.T) {
		authorizerMock := &mockAuthorizer{decision: authorizer.DecisionNoOpinion}
		resolver := &mockResolver{}
		attributes := withArgsAttributes(withoutChildResolverSet)

		clientset := fake.NewSimpleClientset()
		directive := NewRBACDirective(authorizerMock, clientset.Discovery())

		_, err := directive(ctx, 0, resolver.Mock, attributes)

		t.Run("Then authorizer is called with proper attributes", func(t *testing.T) {
			verifyCommonAttributes(t, authorizerMock.lastCallAttributes)
		})
		t.Run("Then next resolver is not called", func(t *testing.T) {
			assert.Equal(t, nil, resolver.lastCallContext)
		})
		t.Run("Then error is returned", func(t *testing.T) {
			assert.NotNil(t, err)
		})
	})
}

type mockAuthorizer struct {
	authorizer.Authorizer

	decision           authorizer.Decision
	lastCallAttributes authorizer.Attributes
}

func (a *mockAuthorizer) Authorize(ctx context.Context, attrs authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
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
