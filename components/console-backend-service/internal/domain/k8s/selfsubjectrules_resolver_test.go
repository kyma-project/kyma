package k8s_test

import (
	"encoding/json"
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
		expected := fixExampleSSRRGQLResponse()

		in := fixExampleSSRRServiceInput("foo")
		inBytes, err := json.Marshal(in)

		out := fixExampleSSRRServiceOutput()

		mockedService := automock.NewSelfSubjectRulesSvc()
		mockedService.On("Create", nil, inBytes).Return(out, nil).Once()
		defer mockedService.AssertExpectations(t)

		converter := automock.NewSelfSubjectRulesConverter()

		converter.On("ToBytes", in).Return(inBytes, nil).Once()
		converter.On("ToGQL", out).Return(expected, nil).Once()

		defer converter.AssertExpectations(t)

		resolver := k8s.NewSelfSubjectRulesResolver(mockedService)
		resolver.SetSelfSubjectRulesConverter(converter)

		defaultNamespace := "foo"
		result, err := resolver.SelfSubjectRulesQuery(nil, &defaultNamespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)

	})

	t.Run("Success - Default Namespace", func(t *testing.T) {
		expected := fixExampleSSRRGQLResponse()

		in := fixExampleSSRRServiceInput("*")
		inBytes, err := json.Marshal(in)

		out := fixExampleSSRRServiceOutput()

		mockedService := automock.NewSelfSubjectRulesSvc()
		mockedService.On("Create", nil, inBytes).Return(out, nil).Once()
		defer mockedService.AssertExpectations(t)

		converter := automock.NewSelfSubjectRulesConverter()
		converter.On("ToBytes", in).Return(inBytes, nil).Once()
		converter.On("ToGQL", out).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSelfSubjectRulesResolver(mockedService)
		resolver.SetSelfSubjectRulesConverter(converter)

		result, err := resolver.SelfSubjectRulesQuery(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func fixExampleSSRRGQLResponse() []gqlschema.ResourceRule {
	return []gqlschema.ResourceRule{
		gqlschema.ResourceRule{
			Verbs:     []string{"a", "b"},
			Resources: []string{"resA", "resB"},
			APIGroups: []string{"gA", "gB"},
		},
	}
}

func fixExampleSSRRServiceInput(namespace string) *authv1.SelfSubjectRulesReview {
	return &authv1.SelfSubjectRulesReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SelfSubjectRulesReview",
			APIVersion: "authorization.k8s.io/v1",
		},
		Spec: authv1.SelfSubjectRulesReviewSpec{
			Namespace: namespace,
		},
	}
}

func fixExampleSSRRServiceOutput() *authv1.SelfSubjectRulesReview {
	return &authv1.SelfSubjectRulesReview{
		Status: authv1.SubjectRulesReviewStatus{
			ResourceRules: []authv1.ResourceRule{
				authv1.ResourceRule{
					Verbs:     []string{"a", "b"},
					Resources: []string{"resA", "resB"},
					APIGroups: []string{"gA", "gB"},
				},
			},
		},
	}
}
