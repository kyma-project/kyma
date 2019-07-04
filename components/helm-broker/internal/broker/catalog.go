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
	bundles, err := svc.finder.FindAll(osbCtx.BrokerNamespace)
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
			Free:        bPlan.Free,
		}

		if len(bPlan.Schemas) > 0 {
			sPlan.Schemas = f.mapToParametersSchemas(bPlan.Schemas)
		}

		sPlans = append(sPlans, sPlan)
	}

	var sTags []string
	for _, tag := range b.Tags {
		sTags = append(sTags, string(tag))
	}

	meta := f.applyOverridesOnBundleMetadata(b.Metadata)

	return osb.Service{
		ID:                  string(b.ID),
		Name:                string(b.Name),
		Description:         b.Description,
		Bindable:            b.Bindable,
		BindingsRetrievable: b.BindingsRetrievable,
		Requires:            b.Requires,
		PlanUpdatable:       b.PlanUpdatable,
		Plans:               sPlans,
		Metadata:            meta.ToMap(),
		Tags:                sTags,
	}, nil
}

func (f *bundleToServiceConverter) mapToParametersSchemas(planSchemas map[internal.PlanSchemaType]internal.PlanSchema) *osb.Schemas {
	ensureServiceInstancesInit := func(in *osb.ServiceInstanceSchema) *osb.ServiceInstanceSchema {
		if in == nil {
			return &osb.ServiceInstanceSchema{}
		}
		return in
	}

	out := &osb.Schemas{}

	if schema, exists := planSchemas[internal.SchemaTypeProvision]; exists {
		out.ServiceInstance = ensureServiceInstancesInit(out.ServiceInstance)
		out.ServiceInstance.Create = &osb.InputParametersSchema{
			Parameters: schema,
		}
	}
	if schema, exists := planSchemas[internal.SchemaTypeUpdate]; exists {
		out.ServiceInstance = ensureServiceInstancesInit(out.ServiceInstance)
		out.ServiceInstance.Update = &osb.InputParametersSchema{
			Parameters: schema,
		}
	}
	if schema, exists := planSchemas[internal.SchemaTypeBind]; exists {
		out.ServiceBinding = &osb.ServiceBindingSchema{
			Create: &osb.RequestResponseSchema{
				InputParametersSchema: osb.InputParametersSchema{
					Parameters: schema,
				},
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
	if m.ProvisionOnlyOnce {
		metaCopy.Labels["provisionOnlyOnce"] = "true"
	}

	return metaCopy
}
