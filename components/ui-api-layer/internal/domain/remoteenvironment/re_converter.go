package remoteenvironment

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type remoteEnvironmentConverter struct{}

func (c *remoteEnvironmentConverter) ToGQL(in *v1alpha1.RemoteEnvironment) gqlschema.RemoteEnvironment {
	if in == nil {
		return gqlschema.RemoteEnvironment{}
	}

	var reServices []gqlschema.RemoteEnvironmentService

	for _, svc := range in.Spec.Services {
		dmSvc := gqlschema.RemoteEnvironmentService{
			ID:                  svc.ID,
			DisplayName:         svc.DisplayName,
			LongDescription:     svc.LongDescription,
			ProviderDisplayName: svc.ProviderDisplayName,
			Tags:                svc.Tags,
			Entries:             c.mapEntriesCRToDTO(svc.Entries),
		}

		reServices = append(reServices, dmSvc)
	}

	dto := gqlschema.RemoteEnvironment{
		Name:        in.Name,
		Description: in.Spec.Description,
		Source: gqlschema.RemoteEnvironmentSource{
			Type:        in.Spec.Source.Type,
			Namespace:   in.Spec.Source.Namespace,
			Environment: in.Spec.Source.Environment,
		},
		Services: reServices,
	}

	return dto
}

func (mapper *remoteEnvironmentConverter) mapEntriesCRToDTO(entries []v1alpha1.Entry) []gqlschema.RemoteEnvironmentEntry {
	dtos := make([]gqlschema.RemoteEnvironmentEntry, 0, len(entries))
	for _, entry := range entries {
		switch entry.Type {
		case "API":
			dtos = append(dtos, gqlschema.RemoteEnvironmentEntry{
				Type:        entry.Type,
				AccessLabel: mapper.ptrString(entry.AccessLabel),
				GatewayUrl:  mapper.ptrString(entry.GatewayUrl),
			})
		case "Events":
			dtos = append(dtos, gqlschema.RemoteEnvironmentEntry{
				Type: entry.Type,
			})
		}
	}
	return dtos
}

// ptrString returns a pointer to the string value passed in.
func (*remoteEnvironmentConverter) ptrString(v string) *string {
	return &v
}
