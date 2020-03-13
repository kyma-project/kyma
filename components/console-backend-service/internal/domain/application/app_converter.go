package application

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type applicationConverter struct{}

func (c *applicationConverter) ToGQL(in *v1alpha1.Application) gqlschema.Application {
	if in == nil {
		return gqlschema.Application{}
	}

	var appServices []gqlschema.ApplicationService

	for _, svc := range in.Spec.Services {
		dmSvc := gqlschema.ApplicationService{
			ID:                  svc.ID,
			DisplayName:         svc.DisplayName,
			LongDescription:     svc.LongDescription,
			ProviderDisplayName: svc.ProviderDisplayName,
			Tags:                svc.Tags,
			Entries:             c.mapEntriesCRToDTO(svc.Entries),
		}

		appServices = append(appServices, dmSvc)
	}

	compassMetadata := gqlschema.CompassMetadata{}

	if in.Spec.CompassMetadata != nil {
		compassMetadata.ApplicationID = in.Spec.CompassMetadata.ApplicationID
	}

	dto := gqlschema.Application{
		Name:            in.Name,
		Labels:          in.Spec.Labels,
		Description:     in.Spec.Description,
		Services:        appServices,
		CompassMetadata: compassMetadata,
	}

	return dto
}

func (c *applicationConverter) mapEntriesCRToDTO(entries []v1alpha1.Entry) []gqlschema.ApplicationEntry {
	dtos := make([]gqlschema.ApplicationEntry, 0, len(entries))
	for _, entry := range entries {
		switch entry.Type {
		case "API":
			dtos = append(dtos, gqlschema.ApplicationEntry{
				Type:        entry.Type,
				AccessLabel: c.ptrString(entry.AccessLabel),
				GatewayURL:  c.ptrString(entry.GatewayUrl),
			})
		case "Events":
			dtos = append(dtos, gqlschema.ApplicationEntry{
				Type: entry.Type,
			})
		}
	}
	return dtos
}

// ptrString returns a pointer to the string value passed in.
func (*applicationConverter) ptrString(v string) *string {
	return &v
}
