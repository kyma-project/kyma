package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRelfSubjectRulesConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &selfSubjectRulesConverter{}
		in := fixExampleSelfSubjectRulesReview()
		result, err := converter.ToGQL(in)
		require.NoError(t, err)

		expected := []gqlschema.ResourceRule{
			gqlschema.ResourceRule{
				Verbs: []string{
					"foo", "bar",
				},
				Resources: []string{
					"resourceA", "resourceB",
				},
				APIGroups: []string{
					"groupA", "groupB",
				},
			},
		}

		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &selfSubjectRulesConverter{}
		expected := []gqlschema.ResourceRule{}
		result, err := converter.ToGQL(&authv1.SelfSubjectRulesReview{})
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &selfSubjectRulesConverter{}
		result, err := converter.ToGQL(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func fixExampleSelfSubjectRulesReview() *authv1.SelfSubjectRulesReview {
	return &authv1.SelfSubjectRulesReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "",
		},
		Status: authv1.SubjectRulesReviewStatus{
			ResourceRules: []authv1.ResourceRule{
				authv1.ResourceRule{
					Verbs: []string{
						"foo", "bar",
					},
					Resources: []string{
						"resourceA", "resourceB",
					},
					APIGroups: []string{
						"groupA", "groupB",
					},
				},
			},
		},
	}
}
