package broker

import (
	"context"

	"github.com/kyma-project/kyma/components/helm-broker/internal"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
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
func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx OsbContext) (*osb.CatalogResponse, error) {
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
	var sPlans []osb.Plan
	for _, bPlan := range b.Plans {
		sPlan := osb.Plan{
			ID:          string(bPlan.ID),
			Name:        string(bPlan.Name),
			Bindable:    bPlan.Bindable,
			Description: bPlan.Description,
			Metadata:    bPlan.Metadata.ToMap(),
		}

		if len(bPlan.Schemas) > 0 {
			sPlan.ParameterSchemas = f.mapToParametersSchemas(bPlan.Schemas)
		}

		sPlans = append(sPlans, sPlan)
	}

	var sTags []string
	for _, tag := range b.Tags {
		sTags = append(sTags, string(tag))
	}

	meta := f.applyOverridesOnBundleMetadata(b.Metadata)

	return osb.Service{
		ID:          string(b.ID),
		Name:        string(b.Name),
		Description: b.Description,
		Bindable:    b.Bindable,
		Plans:       sPlans,
		Metadata:    meta.ToMap(),
		Tags:        sTags,
	}, nil
}

func (f *bundleToServiceConverter) mapToParametersSchemas(planSchemas map[internal.PlanSchemaType]internal.PlanSchema) *osb.ParameterSchemas {
	ensureServiceInstancesInit := func(in *osb.ServiceInstanceSchema) *osb.ServiceInstanceSchema {
		if in == nil {
			return &osb.ServiceInstanceSchema{}
		}
		return in
	}

	out := &osb.ParameterSchemas{}

	if schema, exists := planSchemas[internal.SchemaTypeProvision]; exists {
		out.ServiceInstances = ensureServiceInstancesInit(out.ServiceInstances)
		out.ServiceInstances.Create = &osb.InputParameters{
			Parameters: schema,
		}
	}
	if schema, exists := planSchemas[internal.SchemaTypeUpdate]; exists {
		out.ServiceInstances = ensureServiceInstancesInit(out.ServiceInstances)
		out.ServiceInstances.Update = &osb.InputParameters{
			Parameters: schema,
		}
	}
	if schema, exists := planSchemas[internal.SchemaTypeBind]; exists {
		out.ServiceBindings = &osb.ServiceBindingSchema{
			Create: &osb.InputParameters{
				Parameters: schema,
			},
		}
	}

	return out
}

func (f *bundleToServiceConverter) applyOverridesOnBundleMetadata(m internal.BundleMetadata) internal.BundleMetadata {
	metaCopy := m.DeepCopy()

	if metaCopy.Labels == nil {
		metaCopy.Labels = map[string]string{}
	}
	// Business requirement that helm bundles are always treated as local
	metaCopy.Labels["local"] = "true"

	return metaCopy
}
