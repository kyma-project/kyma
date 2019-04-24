package k8s_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSelfSubjectRulesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := []*gqlschema.ResourceRule{}

		in := &authv1.SelfSubjectRulesReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SelfSubjectRulesReview",
				APIVersion: "authorization.k8s.io/v1",
			},
			Spec: authv1.SelfSubjectRulesReviewSpec{
				Namespace: "foo",
			},
		}
		out := &authv1.SelfSubjectRulesReview{}

		mockedService := automock.NewSelfSubjectRulesSvc()
		mockedService.On("Create", nil, in).Return(out, nil).Once()
		defer mockedService.AssertExpectations(t)

		converter := automock.NewSelfSubjectRulesConverter()
		converter.On("ToGQL", out).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSelfSubjectRulesResolver(mockedService)
		resolver.SetSelfSubjectRulesConverter(converter)

		defaultnamespace := "foo"
		result, err := resolver.SelfSubjectRulesQuery(nil, &defaultnamespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
