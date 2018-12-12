package broker

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

//go:generate mockery -name=converter -output=automock -outpkg=automock -case=underscore
type converter interface {
	Convert(name internal.RemoteEnvironmentName, svc internal.Service) (osb.Service, error)
}

//go:generate mockery -name=reEnabledChecker -output=automock -outpkg=automock -case=underscore
type reEnabledChecker interface {
	IsRemoteEnvironmentEnabled(namespace, name string) (bool, error)
}

type catalogService struct {
	finder           remoteEnvironmentFinder
	reEnabledChecker reEnabledChecker
	conv             converter
}

func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx osbContext) (*osb.CatalogResponse, error) {
	reList, err := svc.finder.FindAll()
	if err != nil {
		return nil, errors.Wrap(err, "while finding Remote Environments")
	}

	resp := osb.CatalogResponse{}
	resp.Services = make([]osb.Service, 0)
	for _, re := range reList {
		enabled, err := svc.reEnabledChecker.IsRemoteEnvironmentEnabled(osbCtx.BrokerNamespace, string(re.Name))
		if err != nil {
			return nil, errors.Wrap(err, "while checking if RE is enabled")
		}
		if !enabled {
			continue
		}

		for _, reSvc := range re.Services {
			s, err := svc.conv.Convert(re.Name, reSvc)
			if err != nil {
				return nil, errors.Wrap(err, "while converting bundle to service")
			}
			resp.Services = append(resp.Services, s)
		}

	}
	return &resp, nil
}

const (
	defaultPlanName        = "default"
	defaultDisplayPlanName = "Default"
	defaultPlanDescription = "Default plan"
)

type reToServiceConverter struct{}

func (c *reToServiceConverter) Convert(name internal.RemoteEnvironmentName, svc internal.Service) (osb.Service, error) {
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

func (c *reToServiceConverter) osbMetadata(name internal.RemoteEnvironmentName, svc internal.Service) (map[string]interface{}, error) {
	metadata := map[string]interface{}{
		"displayName":                svc.DisplayName,
		"providerDisplayName":        svc.ProviderDisplayName,
		"longDescription":            svc.LongDescription,
		"remoteEnvironmentServiceId": string(svc.ID),
		"labels":                     svc.Labels,
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
func (*reToServiceConverter) isSvcBindable(svc internal.Service) bool {
	return svc.APIEntry != nil
}

func (*reToServiceConverter) osbPlans(svcID internal.RemoteServiceID) []osb.Plan {
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

func (*reToServiceConverter) buildBindingLabels(accLabel string) (map[string]string, error) {
	if accLabel == "" {
		return nil, errors.New("accessLabel field is required to build bindingLabels")
	}
	bindingLabels := make(map[string]string)

	// value is set to true to ensure backward compatibility
	bindingLabels[accLabel] = "true"

	return bindingLabels, nil
}
