package syncer

import (
	"encoding/json"
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppCRMapperV2ToModel(t *testing.T) {
	// given
	schema := `{
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "string",
              "title": "ProvisionSchema"
            }`

	fix := v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app",
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:         "EC description",
			DisplayName:         "DisplayName",
			ProviderDisplayName: "ProviderDisplayName",
			LongDescription:     "LongDescription",
			Labels:              map[string]string{"k": "v", "k2": "v2"},
			Tags:                []string{"tag", "tag2"},
			CompassMetadata: &v1alpha1.CompassMetadata{
				ApplicationID: "123",
			},
			Services: []v1alpha1.Service{
				{
					ID:          "123",
					Name:        "name",
					DisplayName: "DisplayName",
					Description: "Description",
					Entries: []v1alpha1.Entry{
						{
							Type:       "API",
							GatewayUrl: "GatewayUrl",
							TargetUrl:  "TargetUrl",
							Name:       "Name",
						},
						{
							Type: "Events",
						},
					},
					AuthCreateParameterSchema: &schema,
				},
			},
		},
	}

	mapper := &appCRMapperV2{}

	// when
	dm, err := mapper.ToModel(&fix)

	// then
	assert.NoError(t, err)

	assert.EqualValues(t, dm.Name, fix.Name)
	assert.EqualValues(t, dm.Description, fix.Spec.Description)
	assert.EqualValues(t, dm.DisplayName, fix.Spec.DisplayName)
	assert.EqualValues(t, dm.ProviderDisplayName, fix.Spec.ProviderDisplayName)
	assert.EqualValues(t, dm.LongDescription, fix.Spec.LongDescription)
	assert.EqualValues(t, dm.CompassMetadata.ApplicationID, fix.Spec.CompassMetadata.ApplicationID)
	assert.EqualValues(t, dm.Labels, fix.Spec.Labels)
	assert.EqualValues(t, dm.Tags, fix.Spec.Tags)

	//	Services
	require.Len(t, dm.Services, 1)
	assert.EqualValues(t, dm.Services[0].ID, fix.Spec.Services[0].ID)
	assert.Equal(t, dm.Services[0].Name, fix.Spec.Services[0].Name)
	assert.Equal(t, dm.Services[0].DisplayName, fix.Spec.Services[0].DisplayName)
	assert.Equal(t, dm.Services[0].Description, fix.Spec.Services[0].Description)
	assert.True(t, dm.Services[0].EventProvider)

	d, err := json.Marshal(dm.Services[0].ServiceInstanceCreateParameterSchema)
	require.NoError(t, err)
	assert.JSONEq(t, string(d), *fix.Spec.Services[0].AuthCreateParameterSchema)

	// Entries
	assert.Len(t, dm.Services[0].Entries, 2)
	assert.Equal(t, dm.Services[0].Entries[0].Type, fix.Spec.Services[0].Entries[0].Type)
	assert.Equal(t, dm.Services[0].Entries[0].Name, fix.Spec.Services[0].Entries[0].Name)
	assert.Equal(t, dm.Services[0].Entries[0].ID, fix.Spec.Services[0].Entries[0].ID)
	assert.Equal(t, dm.Services[0].Entries[0].TargetURL, fix.Spec.Services[0].Entries[0].TargetUrl)

	assert.Equal(t, dm.Services[0].Entries[1].Type, fix.Spec.Services[0].Entries[1].Type)
}

func TestAppCRMapperV2EventProviderTrue(t *testing.T) {
	// given
	mapper := &appCRMapperV2{}

	// when
	dmEvent, errEvent := mapper.ToModel(fixEventsBasedApp())
	dmAPIEvent, errEventAPI := mapper.ToModel(fixAPIAndEventsApp())

	// then
	require.NoError(t, errEvent)
	require.NoError(t, errEventAPI)

	assert.True(t, dmEvent.Services[0].EventProvider)
	assert.True(t, dmAPIEvent.Services[0].EventProvider)
}

func TestAppCRMapperV2EventProviderFalse(t *testing.T) {
	// given
	mapper := &appCRMapperV2{}

	// when
	dmAPI, err := mapper.ToModel(fixAPIBasedApp())

	// then
	require.NoError(t, err)
	assert.False(t, dmAPI.Services[0].EventProvider)
}

func TestAppCRMapperToModel(t *testing.T) {
	// given
	fix := v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app",
		},
		Spec: v1alpha1.ApplicationSpec{
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

	mapper := &appCRMapper{}

	// when
	dm, err := mapper.ToModel(&fix)

	// then
	assert.NoError(t, err)

	assert.Equal(t, dm.Description, fix.Spec.Description)
	assert.Equal(t, dm.Name, internal.ApplicationName(fix.Name))

	require.Len(t, dm.Services, 1)
	assert.Equal(t, string(dm.Services[0].ID), fix.Spec.Services[0].ID)
	assert.Equal(t, dm.Services[0].Tags, fix.Spec.Services[0].Tags)
	assert.Equal(t, dm.Services[0].DisplayName, fix.Spec.Services[0].DisplayName)
	assert.Equal(t, dm.Services[0].LongDescription, fix.Spec.Services[0].LongDescription)
	assert.Equal(t, dm.Services[0].ProviderDisplayName, fix.Spec.Services[0].ProviderDisplayName)

	assert.Len(t, dm.Services[0].Entries, 1)
	assert.Equal(t, dm.Services[0].Entries[0].Type, fix.Spec.Services[0].Entries[0].Type)
	assert.Equal(t, dm.Services[0].Entries[0].AccessLabel, fix.Spec.Services[0].Entries[0].AccessLabel)
	assert.Equal(t, dm.Services[0].Entries[0].GatewayURL, fix.Spec.Services[0].Entries[0].GatewayUrl)
}

func TestEventProviderTrue(t *testing.T) {
	fixEventApp := fixEventsBasedApp()
	fixAPIEventsApp := fixAPIAndEventsApp()

	mapper := &appCRMapper{}

	// when
	dmEvent, errEvent := mapper.ToModel(fixEventApp)
	dmAPIEvent, errEventAPI := mapper.ToModel(fixAPIEventsApp)

	// then
	require.NoError(t, errEvent)
	require.NoError(t, errEventAPI)

	assert.Equal(t, true, dmEvent.Services[0].EventProvider)
	assert.Equal(t, true, dmAPIEvent.Services[0].EventProvider)
}

func TestEventProviderFalse(t *testing.T) {
	mapper := &appCRMapper{}

	// when
	dmAPI, err := mapper.ToModel(fixAPIBasedApp())

	// then
	require.NoError(t, err)
	assert.Equal(t, false, dmAPI.Services[0].EventProvider)
}

func fixEventsBasedApp() *v1alpha1.Application {
	return &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			CompassMetadata: &v1alpha1.CompassMetadata{},
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

func fixAPIBasedApp() *v1alpha1.Application {
	return &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			CompassMetadata: &v1alpha1.CompassMetadata{},
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

func fixAPIAndEventsApp() *v1alpha1.Application {
	return &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			CompassMetadata: &v1alpha1.CompassMetadata{},
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
