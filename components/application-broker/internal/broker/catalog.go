package broker

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

//go:generate mockery -name=converter -output=automock -outpkg=automock -case=underscore
type converter interface {
	Convert(name internal.ApplicationName, svc internal.Service) (osb.Service, error)
}

//go:generate mockery -name=serviceCheckerFactory -output=automock -outpkg=automock -case=underscore
type serviceCheckerFactory interface {
	NewServiceChecker(namespace, name string) (access.ServiceEnabledChecker, error)
}

type catalogService struct {
	finder            applicationFinder
	appEnabledChecker serviceCheckerFactory
	conv              converter
	mockedV2          bool
}

func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx osbContext) (*osb.CatalogResponse, error) {
	resp := osb.CatalogResponse{}
	resp.Services = make([]osb.Service, 0)

	if svc.mockedV2 {
		app := Application{}
		err := yaml.Unmarshal([]byte(appYAML), &app)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshaling mocked Application")
		}
		svcChecker, err := svc.appEnabledChecker.NewServiceChecker(osbCtx.BrokerNamespace, app.Name)
		if err != nil {
			return nil, errors.Wrap(err, "while checking if Application is enabled")
		}

		conv := appToServiceConverterV2{svcChecker: svcChecker}
		s, err := conv.Convert(app)
		if err != nil {
			return nil, errors.Wrap(err, "while converting mocked Application")
		}
		resp.Services = append(resp.Services, s)

	} else { // old working approach (tmp solution)
		appList, err := svc.finder.FindAll()
		if err != nil {
			return nil, errors.Wrap(err, "while finding Applications")
		}

		for _, app := range appList {
			svcChecker, err := svc.appEnabledChecker.NewServiceChecker(osbCtx.BrokerNamespace, string(app.Name))
			if err != nil {
				return nil, errors.Wrap(err, "while checking if Application is enabled")
			}

			for _, s := range app.Services {
				if !svcChecker.IsServiceEnabled(s) {
					continue
				}
				s, err := svc.conv.Convert(app.Name, s)
				if err != nil {
					return nil, errors.Wrap(err, "while converting bundle to service")
				}
				resp.Services = append(resp.Services, s)
			}

		}
	}

	return &resp, nil
}

type appToServiceConverterV2 struct {
	svcChecker access.ServiceEnabledChecker
}

func (c *appToServiceConverterV2) Convert(app Application) (osb.Service, error) {
	// service(class) metadata
	svcMetadata := c.toServiceMetadata(app)

	// plans
	plans := c.toPlans(app.Services)

	// service(class)
	return osb.Service{
		ID:          app.CompassMetadata.ApplicationID,
		Name:        app.Name,
		Description: app.Description,
		Bindable:    true,
		Plans:       plans,
		Metadata:    svcMetadata,
	}, nil
}
func (c *appToServiceConverterV2) toServiceMetadata(app Application) map[string]interface{} {
	if app.Labels == nil {
		app.Labels = map[string]string{}
	}
	app.Labels["documentation-per-plan"] = "true"
	// business requirement that services can be always provisioned only once
	app.Labels["provisionOnlyOnce"] = "true"

	return map[string]interface{}{
		"displayName":         app.DisplayName,
		"providerDisplayName": app.ProviderDisplayName,
		"longDescription":     app.LongDescription,
		"labels":              app.Labels,
	}
}

func (c *appToServiceConverterV2) toPlans(services []Service) []osb.Plan {
	var plans []osb.Plan
	for _, svc := range services {
		//if !c.svcChecker.IsServiceEnabled(internal.Service{ID: internal.ApplicationServiceID(svc.ID)}) {
		//	continue
		//}

		plan := osb.Plan{
			ID:          svc.ID,
			Name:        svc.Name,
			Description: svc.Description,
			Metadata: map[string]interface{}{
				"displayName": svc.DisplayName,
			},
			Schemas:  nil,
			Bindable: c.isPlanBindable(svc),
		}
		plans = append(plans, plan)
	}

	return plans
}

func (*appToServiceConverterV2) isPlanBindable(svc Service) *bool {
	ptr := func(b bool) *bool {
		return &b
	}

	for _, e := range svc.Entries {
		if e.Type == "API" {
			return ptr(true)
		}
	}
	return ptr(false)
}

// Deprecated Converter Implementation

const (
	defaultPlanName        = "default"
	defaultDisplayPlanName = "Default"
	defaultPlanDescription = "Default plan"
)

type appToServiceConverter struct{}

func (c *appToServiceConverter) Convert(name internal.ApplicationName, svc internal.Service) (osb.Service, error) {
	metadata, err := c.osbMetadata(name, svc)
	if err != nil {
		return osb.Service{}, errors.Wrap(err, "while creating the metadata object")
	}

	osbService := osb.Service{
		Name:        svc.Name,
		ID:          string(svc.ID),
		Description: svc.Description,
		Bindable:    c.isSvcBindable(svc),
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
	if svc.APIEntry != nil {
		// future: comma separated labels, must be supported on Service API
		bindingLabels, err := c.buildBindingLabels(svc.APIEntry.AccessLabel)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create binding labels")
		}
		metadata["bindingLabels"] = bindingLabels
	}

	return metadata, nil
}

// isSvcBindable checks if service is bindable. If APIEntry is not set then service provides only events,
// so it is not bindable and false is returned
func (*appToServiceConverter) isSvcBindable(svc internal.Service) bool {
	return svc.APIEntry != nil
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
	labels["provisionOnlyOnce"] = "true"

	return labels
}
