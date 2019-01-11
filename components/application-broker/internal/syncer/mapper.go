package syncer

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

type appCRMapper struct{}

const (
	api    = "API"
	events = "Events"
)

// ToModel produces Application domain model from Application custom resource
func (app *appCRMapper) ToModel(dto *v1alpha1.Application) *internal.Application {
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
			APIEntry:            app.extractAPIEntryAsModel(svc.Entries),
			EventProvider:       app.extractEventEntryAsModel(svc.Entries),
		}

		appSvcs = append(appSvcs, dmSvc)
	}

	dm := &internal.Application{
		Name:        internal.ApplicationName(dto.Name),
		Description: dto.Spec.Description,
		Services:    appSvcs,
		AccessLabel: dto.Spec.AccessLabel,
	}

	return dm
}

func (*appCRMapper) extractAPIEntryAsModel(entries []v1alpha1.Entry) *internal.APIEntry {
	for _, entry := range entries {
		switch entry.Type {
		case api:
			// TODO(entry-simplification): this is an accepted simplification until
			// explicit support of many APIEntry and EventEntry.
			// For now we are know that only one entry of type API is allowed,
			// so we are returning immediately
			return &internal.APIEntry{
				Entry:       internal.Entry{Type: entry.Type},
				AccessLabel: entry.AccessLabel,
				GatewayURL:  entry.GatewayUrl,
			}

		}
	}
	return nil
}
func (*appCRMapper) extractEventEntryAsModel(entries []v1alpha1.Entry) bool {
	for _, entry := range entries {
		switch entry.Type {
		case events:
			// TODO(entry-simplification): this is an accepted simplification until
			// explicit support of many APIEntry and EventEntry.
			return true
		}
	}
	return false
}
