package broker

import (
	"context"

	"github.com/pkg/errors"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type catalogService struct {
	finder bundleFinder
	conv   converter
}

//go:generate mockery -name=converter -output=automock -outpkg=automock -case=underscore
type converter interface {
	Convert(b *internal.Bundle) (osb.Service, error)
}

// TODO: switch from osb.CatalogResponse to CatalogSuccessResponseDTO
func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx osbContext) (*osb.CatalogResponse, error) {
	bundles, err := svc.finder.FindAll()
	if err != nil {
		return nil, errors.Wrap(err, "while finding all bundles")
	}

	resp := osb.CatalogResponse{}
	resp.Services = make([]osb.Service, len(bundles))
	for idx, b := range bundles {
		s, err := svc.conv.Convert(b)
		if err != nil {
			return nil, errors.Wrap(err, "while converting bundle to service")
		}
		resp.Services[idx] = s
	}
	return &resp, nil
}

type bundleToServiceConverter struct{}

func (f *bundleToServiceConverter) Convert(b *internal.Bundle) (osb.Service, error) {
	sPlans := make([]osb.Plan, 0)
	for _, bPlan := range b.Plans {

		sPlan := osb.Plan{
			ID:          string(bPlan.ID),
			Name:        string(bPlan.Name),
			Bindable:    bPlan.Bindable,
			Description: bPlan.Description,
			ParameterSchemas: &osb.ParameterSchemas{
				ServiceInstances: &osb.ServiceInstanceSchema{
					Create: &osb.InputParameters{},
					Update: &osb.InputParameters{},
				},
				ServiceBindings: &osb.ServiceBindingSchema{
					Create: &osb.InputParameters{},
				},
			},
			Metadata: bPlan.Metadata.ToMap(),
		}
		if provisionSchema, exists := bPlan.Schemas[internal.SchemaTypeProvision]; exists {
			sPlan.ParameterSchemas.ServiceInstances.Create = &osb.InputParameters{
				Parameters: provisionSchema,
			}
		}
		if updateSchema, exists := bPlan.Schemas[internal.SchemaTypeUpdate]; exists {
			sPlan.ParameterSchemas.ServiceInstances.Update = &osb.InputParameters{
				Parameters: updateSchema,
			}
		}
		if bindSchema, exists := bPlan.Schemas[internal.SchemaTypeBind]; exists {
			sPlan.ParameterSchemas.ServiceBindings.Create = &osb.InputParameters{
				Parameters: bindSchema,
			}
		}

		sPlans = append(sPlans, sPlan)
	}

	var sTags []string
	for _, tag := range b.Tags {
		sTags = append(sTags, string(tag))
	}

	return osb.Service{
		ID:          string(b.ID),
		Name:        string(b.Name),
		Description: b.Description,
		Bindable:    b.Bindable,
		Plans:       sPlans,
		Metadata:    b.Metadata.ToMap(),
		Tags:        sTags,
	}, nil
}
