package broker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
)

//go:generate mockery -name=converter -output=automock -outpkg=automock -case=underscore
type converter interface {
	Convert(name internal.RemoteEnvironmentName, source internal.Source, svc internal.Service) (osb.Service, error)
}

type catalogService struct {
	finder remoteEnvironmentFinder
	conv   converter
}

func (svc *catalogService) GetCatalog(ctx context.Context, osbCtx osbContext) (*osb.CatalogResponse, error) {
	reList, err := svc.finder.FindAll()
	if err != nil {
		return nil, errors.Wrap(err, "while finding Remote Environments")
	}

	resp := osb.CatalogResponse{}
	resp.Services = make([]osb.Service, 0)
	for _, re := range reList {
		for _, reSvc := range re.Services {
			s, err := svc.conv.Convert(re.Name, re.Source, reSvc)
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

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

type reToServiceConverter struct{}

func (c *reToServiceConverter) Convert(name internal.RemoteEnvironmentName, source internal.Source, svc internal.Service) (osb.Service, error) {
	metadata, err := c.osbMetadata(name, source, svc)
	if err != nil {
		return osb.Service{}, errors.Wrap(err, "while creating the metadata object")
	}

	osbService := osb.Service{
		Name:        c.createOsbServiceName(svc.DisplayName, svc.ID),
		ID:          string(svc.ID),
		Description: svc.LongDescription,
		Bindable:    c.isSvcBindable(svc),
		Metadata:    metadata,
		Plans:       c.osbPlans(svc.ID),
		Tags:        svc.Tags,
	}

	return osbService, nil
}

func (c *reToServiceConverter) osbMetadata(name internal.RemoteEnvironmentName, source internal.Source, svc internal.Service) (map[string]interface{}, error) {
	metadata := map[string]interface{}{
		"displayName":                svc.DisplayName,
		"providerDisplayName":        c.osbProviderDisplayName(name, svc),
		"longDescription":            svc.LongDescription,
		"remoteEnvironmentServiceId": string(svc.ID),
		"source": map[string]interface{}{
			"environment": source.Environment,
			"type":        source.Type,
			"namespace":   source.Namespace,
		},
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

// osbProviderDisplayName returns the ProviderDisplayName in a such way that we are able to easily distinguish the remote environment from the same provider
// e.g. instead of having always `3rd Party Provider`, we will have `3rd Party Provider - stage`, `3rd Party Provider - prod-us`
func (*reToServiceConverter) osbProviderDisplayName(name internal.RemoteEnvironmentName, svc internal.Service) string {
	return fmt.Sprintf("%s - %s", svc.ProviderDisplayName, name)
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

// createOsbServiceName creates the OSB Service Name for given RemoteEnvironment Service.
// The OSB Service Name is used in the Service Catalog as the clusterServiceClassExternalName, so it need to be normalized.
//
// Normalization rules:
// - MUST only contain lowercase characters, numbers and hyphens (no spaces).
// - MUST be unique across all service objects returned in this response. MUST be a non-empty string.
func (*reToServiceConverter) createOsbServiceName(name string, id internal.RemoteServiceID) string {
	// generate 5 characters suffix from the id
	sha := sha1.New()
	sha.Write([]byte(id))
	suffix := hex.EncodeToString(sha.Sum(nil))[:5]

	// remove all characters, which is not alpha numeric
	name = nonAlphaNumeric.ReplaceAllString(name, "-")

	// to lower
	name = strings.Map(unicode.ToLower, name)

	// trim dashes if exists
	name = strings.TrimSuffix(name, "-")
	if len(name) > 57 {
		name = name[:57]
	}

	// add suffix
	name = fmt.Sprintf("%s-%s", name, suffix)

	// remove dash prefix if exists
	//  - can happen, if the name was empty before adding suffix empty or had dash prefix
	name = strings.TrimPrefix(name, "-")
	return name
}
