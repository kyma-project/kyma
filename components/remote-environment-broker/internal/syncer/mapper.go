package syncer

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
)

type reCRMapper struct{}

const (
	api    = "API"
	events = "Events"
)

// ToModel produces RemoteEnvironment domain model from RemoteEnvironment custom resource
func (re *reCRMapper) ToModel(dto *v1alpha1.RemoteEnvironment) *internal.RemoteEnvironment {
	var reServices []internal.Service

	for _, svc := range dto.Spec.Services {
		dmSvc := internal.Service{
			ID:                  internal.RemoteServiceID(svc.ID),
			DisplayName:         svc.DisplayName,
			LongDescription:     svc.LongDescription,
			ProviderDisplayName: svc.ProviderDisplayName,
			Tags:                svc.Tags,
			APIEntry:            re.extractAPIEntryAsModel(svc.Entries),
			EventProvider:       re.extractEventEntryAsModel(svc.Entries),
		}

		reServices = append(reServices, dmSvc)
	}

	dm := &internal.RemoteEnvironment{
		Name:        internal.RemoteEnvironmentName(dto.Name),
		Description: dto.Spec.Description,
		Source: internal.Source{
			Type:        dto.Spec.Source.Type,
			Namespace:   dto.Spec.Source.Namespace,
			Environment: dto.Spec.Source.Environment,
		},
		Services:    reServices,
		AccessLabel: dto.Spec.AccessLabel,
	}

	return dm
}

func (*reCRMapper) extractAPIEntryAsModel(entries []v1alpha1.Entry) *internal.APIEntry {
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
func (*reCRMapper) extractEventEntryAsModel(entries []v1alpha1.Entry) bool {
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
