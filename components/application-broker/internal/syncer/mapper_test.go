package syncer

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReCRMapperToModel(t *testing.T) {
	// given
	fix := v1alpha1.RemoteEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "re",
		},
		Spec: v1alpha1.RemoteEnvironmentSpec{
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
							Type: "Events",
						},
					},
				},
			},
		},
	}

	mapper := &reCRMapper{}

	// when
	dm := mapper.ToModel(&fix)

	// then
	assert.Equal(t, dm.Description, fix.Spec.Description)
	assert.Equal(t, dm.Name, internal.RemoteEnvironmentName(fix.Name))

	require.Len(t, dm.Services, 1)
	assert.Equal(t, string(dm.Services[0].ID), fix.Spec.Services[0].ID)
	assert.Equal(t, dm.Services[0].Tags, fix.Spec.Services[0].Tags)
	assert.Equal(t, dm.Services[0].DisplayName, fix.Spec.Services[0].DisplayName)
	assert.Equal(t, dm.Services[0].LongDescription, fix.Spec.Services[0].LongDescription)
	assert.Equal(t, dm.Services[0].ProviderDisplayName, fix.Spec.Services[0].ProviderDisplayName)

	assert.Equal(t, dm.Services[0].APIEntry.Type, fix.Spec.Services[0].Entries[0].Type)
	assert.Equal(t, dm.Services[0].APIEntry.AccessLabel, fix.Spec.Services[0].Entries[0].AccessLabel)
	assert.Equal(t, dm.Services[0].APIEntry.GatewayURL, fix.Spec.Services[0].Entries[0].GatewayUrl)
}

func TestEventProviderTrue(t *testing.T) {
	fixEventRE := fixEventsBasedRE()
	fixAPIEventsRE := fixAPIAndEventsRE()

	mapper := &reCRMapper{}

	// when
	dmEvent := mapper.ToModel(fixEventRE)
	dmAPIEvent := mapper.ToModel(fixAPIEventsRE)

	// then
	assert.Equal(t, true, dmEvent.Services[0].EventProvider)
	assert.Equal(t, true, dmAPIEvent.Services[0].EventProvider)
}

func TestEventProviderFalse(t *testing.T) {
	mapper := &reCRMapper{}

	// when
	dmAPI := mapper.ToModel(fixAPIBasedRE())

	// then
	assert.Equal(t, false, dmAPI.Services[0].EventProvider)
}

func fixEventsBasedRE() *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services: []v1alpha1.Service{
				{
					ID: "123",
					Entries: []v1alpha1.Entry{
						{
							Type: "Events",
						},
					},
				},
			},
		},
	}
}

func fixAPIBasedRE() *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services: []v1alpha1.Service{
				{
					ID: "123",
					Entries: []v1alpha1.Entry{
						{
							Type:        "API",
							GatewayUrl:  "url",
							AccessLabel: "label",
						},
					},
				},
			},
		},
	}
}

func fixAPIAndEventsRE() *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services: []v1alpha1.Service{
				{
					ID: "123",
					Entries: []v1alpha1.Entry{
						{
							Type:        "API",
							GatewayUrl:  "url",
							AccessLabel: "label",
						},
						{
							Type: "Events",
						},
					},
				},
			},
		},
	}
}
