package broker

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

//go:generate mockery -name=converter -output=automock -outpkg=automock -case=underscore
type converter interface {
	Convert(svcChecker access.ServiceEnabledChecker, app internal.Application) ([]osb.Service, error)
}

//go:generate mockery -name=serviceCheckerFactory -output=automock -outpkg=automock -case=underscore
type serviceCheckerFactory interface {
	NewServiceChecker(namespace, name string) (access.ServiceEnabledChecker, error)
}

type catalogService struct {
	finder            applicationFinder
	appEnabledChecker serviceCheckerFactory
	conv              converter
}

func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx osbContext) (*osb.CatalogResponse, error) {
	resp := osb.CatalogResponse{}

	appList, err := svc.finder.FindAll()
	if err != nil {
		return nil, errors.Wrap(err, "while finding Applications")
	}

	for _, app := range appList {
		svcChecker, err := svc.appEnabledChecker.NewServiceChecker(osbCtx.BrokerNamespace, string(app.Name))
		if err != nil {
			return nil, errors.Wrap(err, "while checking if Application is enabled")
		}

		s, err := svc.conv.Convert(svcChecker, *app)
		if err != nil {
			return nil, errors.Wrap(err, "while converting application to OSB services")
		}
		resp.Services = append(resp.Services, s...)
	}

	return &resp, nil
}

const (
	documentationPerPlanLabelName = "documentation-per-plan"
	provisionOnlyOnceLabelName    = "provisionOnlyOnce"
)

type appToServiceConverterV2 struct{}

func (c *appToServiceConverterV2) Convert(svcChecker access.ServiceEnabledChecker, app internal.Application) ([]osb.Service, error) {
	// plans
	plans := c.toPlans(svcChecker, app.Services)
	if len(plans) == 0 {
		return nil, errors.Errorf("None plans were mapped from Application Services. Used Checker: %s, Services: [%+v].", svcChecker.IdentifyYourself(), app.Services)
	}

	// service(class) metadata
	svcMetadata := c.toServiceMetadata(app)

	// service(class)
	return []osb.Service{
		{
			ID:          app.CompassMetadata.ApplicationID,
			Name:        string(app.Name),
			Description: app.Description,
			Bindable:    true,
			Plans:       plans,
			Metadata:    svcMetadata,
			Tags:        app.Tags,
		},
	}, nil
}
func (c *appToServiceConverterV2) toServiceMetadata(app internal.Application) map[string]interface{} {
	if app.Labels == nil {
		app.Labels = map[string]string{}
	}

	// In new approach documentation is uploaded per plan and not per class
	app.Labels[documentationPerPlanLabelName] = "true"

	return map[string]interface{}{
		"displayName":         app.DisplayName,
		"providerDisplayName": app.ProviderDisplayName,
		"longDescription":     app.LongDescription,
		"labels":              app.Labels,
	}
}

func (c *appToServiceConverterV2) toPlans(svcChecker access.ServiceEnabledChecker, services []internal.Service) []osb.Plan {
	var plans []osb.Plan
	for _, svc := range services {
		if !svcChecker.IsServiceEnabled(svc) {
			continue
		}

		plan := osb.Plan{
			ID:          string(svc.ID),
			Name:        svc.Name,
			Description: svc.Description,
			Metadata: map[string]interface{}{
				"displayName": svc.DisplayName,
			},
			Schemas:  c.toSchemas(svc),
			Bindable: boolPtr(svc.IsBindable()),
		}
		plans = append(plans, plan)
	}

	return plans
}

func (c *appToServiceConverterV2) toSchemas(svc internal.Service) *osb.Schemas {
	if svc.ServiceInstanceCreateParameterSchema == nil {
		return nil
	}

	return &osb.Schemas{
		ServiceInstance: &osb.ServiceInstanceSchema{
			Create: &osb.InputParametersSchema{
				Parameters: svc.ServiceInstanceCreateParameterSchema,
			},
		},
	}
}

// Deprecated Converter Implementation

const (
	defaultPlanName        = "default"
	defaultDisplayPlanName = "Default"
	defaultPlanDescription = "Default plan"
)

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
type appToServiceConverter struct{}

func (c *appToServiceConverter) Convert(svcChecker access.ServiceEnabledChecker, app internal.Application) ([]osb.Service, error) {
	services := make([]osb.Service, 0)
	for _, s := range app.Services {
		if !svcChecker.IsServiceEnabled(s) {
			continue
		}
		s, err := c.convert(app.Name, s)
		if err != nil {
			return nil, errors.Wrap(err, "while converting application to service")
		}
		services = append(services, s)
	}

	return services, nil
}

func (c *appToServiceConverter) convert(name internal.ApplicationName, svc internal.Service) (osb.Service, error) {
	metadata, err := c.osbMetadata(name, svc)
	if err != nil {
		return osb.Service{}, errors.Wrap(err, "while creating the metadata object")
	}

	osbService := osb.Service{
		Name:        svc.Name,
		ID:          string(svc.ID),
		Description: svc.Description,
		Bindable:    svc.IsBindable(),
		Metadata:    metadata,
		Plans:       c.osbPlans(svc.ID),
		Tags:        svc.Tags,
	}

	return osbService, nil
}

func (c *appToServiceConverter) osbMetadata(name internal.ApplicationName, svc internal.Service) (map[string]interface{}, error) {
	metadata := map[string]interface{}{
		"displayName":          svc.DisplayName,
		"providerDisplayName":  svc.ProviderDisplayName,
		"longDescription":      svc.LongDescription,
		"applicationServiceId": string(svc.ID),
		"labels":               c.applyOverridesOnLabels(svc.Labels),
	}

	// TODO(entry-simplification): this is an accepted simplification until
	// explicit support of many APIEntry and EventEntry
	if len(svc.Entries) == 1 && svc.Entries[0].APIEntry != nil {
		// future: comma separated labels, must be supported on Service API
		bindingLabels, err := c.buildBindingLabels(svc.Entries[0].APIEntry.AccessLabel)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create binding labels")
		}
		metadata["bindingLabels"] = bindingLabels
	}

	return metadata, nil
}

func (*appToServiceConverter) osbPlans(svcID internal.ApplicationServiceID) []osb.Plan {
	plan := osb.Plan{
		ID:          fmt.Sprintf("%s-plan", svcID),
		Name:        defaultPlanName,
		Description: defaultPlanDescription,
		Metadata: map[string]interface{}{
			"displayName": defaultDisplayPlanName,
		},
	}

	return []osb.Plan{plan}
}

func (*appToServiceConverter) buildBindingLabels(accLabel string) (map[string]string, error) {
	if accLabel == "" {
		return nil, errors.New("accessLabel field is required to build bindingLabels")
	}
	bindingLabels := make(map[string]string)

	// value is set to true to ensure backward compatibility
	bindingLabels[accLabel] = "true"

	return bindingLabels, nil
}

func (*appToServiceConverter) applyOverridesOnLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	// business requirement that services can be always provisioned only once
	labels[provisionOnlyOnceLabelName] = "true"

	return labels
}

func boolPtr(in bool) *bool {
	return &in
}
