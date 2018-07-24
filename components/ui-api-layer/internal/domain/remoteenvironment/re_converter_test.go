package remoteenvironment

import (
	"testing"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoteEnvironmentConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		fix := v1alpha1.RemoteEnvironment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "re",
			},
			Spec: v1alpha1.RemoteEnvironmentSpec{
				Source: v1alpha1.Source{
					Environment: "production",
					Type:        "commerce",
					Namespace:   "local.kyma.commerce",
				},
				Description: "EC description",
				Services: []v1alpha1.Service{
					{
						ID:                  "123",
						DisplayName:         "name",
						Tags:                []string{"tag1", "tag2"},
						LongDescription:     "desc",
						ProviderDisplayName: "name",
						Entries: []v1alpha1.Entry{
							{
								Type:        "API",
								GatewayUrl:  "url",
								AccessLabel: "label",
							},
							{
								Type: "Event",
							},
						},
					},
				},
			},
		}

		converter := &remoteEnvironmentConverter{}

		// when
		dto := converter.ToGQL(&fix)

		// then
		assert.Equal(t, dto.Description, fix.Spec.Description)
		assert.Equal(t, dto.Name, fix.Name)

		assert.Equal(t, dto.Source.Type, fix.Spec.Source.Type)
		assert.Equal(t, dto.Source.Environment, fix.Spec.Source.Environment)
		assert.Equal(t, dto.Source.Namespace, fix.Spec.Source.Namespace)

		require.Len(t, dto.Services, 1)
		assert.Equal(t, dto.Services[0].ID, fix.Spec.Services[0].ID)
		assert.Equal(t, dto.Services[0].Tags, fix.Spec.Services[0].Tags)
		assert.Equal(t, dto.Services[0].DisplayName, fix.Spec.Services[0].DisplayName)
		assert.Equal(t, dto.Services[0].LongDescription, fix.Spec.Services[0].LongDescription)
		assert.Equal(t, dto.Services[0].ProviderDisplayName, fix.Spec.Services[0].ProviderDisplayName)

		require.Len(t, dto.Services[0].Entries, 1)
		assert.Equal(t, dto.Services[0].Entries[0].Type, fix.Spec.Services[0].Entries[0].Type)
		assert.Equal(t, dto.Services[0].Entries[0].AccessLabel, &fix.Spec.Services[0].Entries[0].AccessLabel)
		assert.Equal(t, dto.Services[0].Entries[0].GatewayUrl, &fix.Spec.Services[0].Entries[0].GatewayUrl)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &remoteEnvironmentConverter{}
		result := converter.ToGQL(&v1alpha1.RemoteEnvironment{})

		assert.Empty(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &remoteEnvironmentConverter{}
		result := converter.ToGQL(nil)

		assert.Empty(t, result)
	})
}
