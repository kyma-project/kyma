package broker

import (
	"context"

	"github.com/kyma-project/kyma/components/helm-broker/internal"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type catalogService struct {
	finder addonFinder
	conv   converter
}

//go:generate mockery -name=converter -output=automock -outpkg=automock -case=underscore
type converter interface {
	Convert(b *internal.Addon) (osb.Service, error)
}

// TODO: switch from osb.CatalogResponse to CatalogSuccessResponseDTO
func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx OsbContext) (*osb.CatalogResponse, error) {
	addons, err := svc.finder.FindAll(osbCtx.BrokerNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "while finding all addons")
	}

	resp := osb.CatalogResponse{}
	resp.Services = make([]osb.Service, len(addons))
	for idx, b := range addons {
		s, err := svc.conv.Convert(b)
		if err != nil {
			return nil, errors.Wrap(err, "while converting addon to service")
		}
		resp.Services[idx] = s
	}
	return &resp, nil
}

type addonToServiceConverter struct{}

func (f *addonToServiceConverter) Convert(addon *internal.Addon) (osb.Service, error) {
	var sPlans []osb.Plan
	for _, bPlan := range addon.Plans {
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
	for _, tag := range addon.Tags {
		sTags = append(sTags, string(tag))
	}

	meta := f.applyOverridesOnAddonMetadata(addon.Metadata)

	return osb.Service{
		ID:                  string(addon.ID),
		Name:                string(addon.Name),
		Description:         addon.Description,
		Bindable:            addon.Bindable,
		BindingsRetrievable: addon.BindingsRetrievable,
		Requires:            addon.Requires,
		PlanUpdatable:       addon.PlanUpdatable,
		Plans:               sPlans,
		Metadata:            meta.ToMap(),
		Tags:                sTags,
	}, nil
}

func (f *addonToServiceConverter) mapToParametersSchemas(planSchemas map[internal.PlanSchemaType]internal.PlanSchema) *osb.Schemas {
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

func (f *addonToServiceConverter) applyOverridesOnAddonMetadata(m internal.AddonMetadata) internal.AddonMetadata {
	metaCopy := m.DeepCopy()

	if metaCopy.Labels == nil {
		metaCopy.Labels = map[string]string{}
	}
	// Business requirement that helm addons are always treated as local
	metaCopy.Labels["local"] = "true"
	if m.ProvisionOnlyOnce {
		metaCopy.Labels["provisionOnlyOnce"] = "true"
	}

	return metaCopy
}
