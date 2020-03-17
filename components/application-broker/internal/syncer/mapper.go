package syncer

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
)

type appCRMapperV2 struct{}

// ToModel produces Application domain model from Application custom resource
func (m *appCRMapperV2) ToModel(app *v1alpha1.Application) (*internal.Application, error) {
	var appSvcs []internal.Service

	for _, svc := range app.Spec.Services {
		instanceSchema, err := m.schema(svc.AuthCreateParameterSchema)
		if err != nil {
			return nil, errors.Wrap(err, "while converting instance create schema")
		}
		dmSvc := internal.Service{
			ID:                                   internal.ApplicationServiceID(svc.ID),
			Name:                                 svc.Name,
			DisplayName:                          svc.DisplayName,
			Description:                          svc.Description,
			Entries:                              m.entriesToModel(svc.Entries),
			EventProvider:                        m.isEventProvider(svc.Entries),
			ServiceInstanceCreateParameterSchema: instanceSchema,
		}

		appSvcs = append(appSvcs, dmSvc)
	}

	return &internal.Application{
		Name:                internal.ApplicationName(app.Name),
		Description:         app.Spec.Description,
		Services:            appSvcs,
		DisplayName:         app.Spec.DisplayName,
		ProviderDisplayName: app.Spec.ProviderDisplayName,
		LongDescription:     app.Spec.LongDescription,
		Labels:              app.Spec.Labels,
		Tags:                app.Spec.Tags,
		CompassMetadata: internal.CompassMetadata{
			ApplicationID: app.Spec.CompassMetadata.ApplicationID,
		},
	}, nil
}

func (*appCRMapperV2) entriesToModel(entries []v1alpha1.Entry) []internal.Entry {
	var mapped []internal.Entry
	for _, entry := range entries {
		e := internal.Entry{Type: entry.Type}
		switch entry.Type {
		case internal.APIEntryType: // map entry
			e.APIEntry = &internal.APIEntry{
				Name:      entry.Name,
				TargetURL: entry.TargetUrl,
				ID:        entry.ID,
			}
		case internal.EventEntryType: // nothing to do
		}

		mapped = append(mapped, e)
	}

	return mapped
}
func (*appCRMapperV2) isEventProvider(entries []v1alpha1.Entry) bool {
	for _, entry := range entries {
		if entry.Type == internal.EventEntryType {
			return true
		}
	}
	return false
}

func (m *appCRMapperV2) schema(schema *string) (map[string]interface{}, error) {
	if schema == nil {
		return nil, nil
	}

	var unmarshaled map[string]interface{}
	err := json.Unmarshal([]byte(*schema), &unmarshaled)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling JSON schema: [%v]", *schema)
	}

	return unmarshaled, nil
}

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
type appCRMapper struct{}

// ToModel produces Application domain model from Application custom resource
func (app *appCRMapper) ToModel(dto *v1alpha1.Application) (*internal.Application, error) {
	var appSvcs []internal.Service

	for _, svc := range dto.Spec.Services {
		dmSvc := internal.Service{
			ID:                  internal.ApplicationServiceID(svc.ID),
			Name:                svc.Name,
			DisplayName:         svc.DisplayName,
			Description:         svc.Description,
			LongDescription:     svc.LongDescription,
			ProviderDisplayName: svc.ProviderDisplayName,
			Tags:                svc.Tags,
			Labels:              svc.Labels,
			Entries:             app.entriesToModel(svc.Entries),
			EventProvider:       app.isEventProvider(svc.Entries),
		}

		appSvcs = append(appSvcs, dmSvc)
	}

	dm := &internal.Application{
		Name:        internal.ApplicationName(dto.Name),
		Description: dto.Spec.Description,
		Services:    appSvcs,
		AccessLabel: dto.Spec.AccessLabel,
	}

	return dm, nil
}

func (*appCRMapper) entriesToModel(entries []v1alpha1.Entry) []internal.Entry {
	for _, entry := range entries {
		switch entry.Type {
		case internal.APIEntryType:
			// TODO(entry-simplification): this is an accepted simplification until
			// explicit support of many APIEntry and EventEntry.
			// For now we are know that only one entry of type API is allowed,
			// so we are returning immediately
			return []internal.Entry{
				{
					Type: entry.Type,
					APIEntry: &internal.APIEntry{
						AccessLabel: entry.AccessLabel,
						GatewayURL:  entry.GatewayUrl,
					},
				},
			}
		}
	}

	return nil
}
func (*appCRMapper) isEventProvider(entries []v1alpha1.Entry) bool {
	for _, entry := range entries {
		switch entry.Type {
		case internal.EventEntryType:
			// TODO(entry-simplification): this is an accepted simplification until
			// explicit support of many APIEntry and EventEntry.
			return true
		}
	}
	return false
}
