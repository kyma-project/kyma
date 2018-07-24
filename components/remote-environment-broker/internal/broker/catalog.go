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
	Convert(source *internal.Source, service *internal.Service) (osb.Service, error)
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

		for _, service := range re.Services {
			s, err := svc.conv.Convert(&re.Source, &service)
			if err != nil {
				return nil, errors.Wrap(err, "while converting bundle to service")
			}
			resp.Services = append(resp.Services, s)
		}

	}
	return &resp, nil
}

type reToServiceConverter struct{}

const (
	defaultPlanName        = "default"
	defaultDisplayPlanName = "Default"
	defaultPlanDescription = "Default plan"
)

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

func (f *reToServiceConverter) Convert(source *internal.Source, service *internal.Service) (osb.Service, error) {
	bindable := true

	plan := osb.Plan{
		ID:          fmt.Sprintf("%s-plan", service.ID),
		Name:        defaultPlanName,
		Description: defaultPlanDescription,
		Metadata: map[string]interface{}{
			"displayName": defaultDisplayPlanName,
		},
	}
	metadata := map[string]interface{}{
		"displayName":                service.DisplayName,
		"providerDisplayName":        service.ProviderDisplayName,
		"longDescription":            service.LongDescription,
		"remoteEnvironmentServiceId": string(service.ID),
		"source": map[string]interface{}{
			"environment": source.Environment,
			"type":        source.Type,
			"namespace":   source.Namespace,
		},
	}

	//TODO(entry-simplification): this is an accepted simplification until
	// explicit support of many APIEntry and EventEntry
	if service.APIEntry != nil {
		// future: comma separated labels, must be supported on Service API
		bindingLabels, err := f.buildBindingLabels(service.APIEntry.AccessLabel)
		if err != nil {
			return osb.Service{}, errors.Wrap(err, "cannot create binding labels")
		}
		metadata["bindingLabels"] = bindingLabels
	} else {
		// service provides only events so it is not bindable
		bindable = false
	}

	osbService := osb.Service{
		ID: string(service.ID),
		// Name is converted to clusterServiceClassExternalName, so it need to be normalized.
		// MUST only contain lowercase characters, numbers and hyphens (no spaces).
		// MUST be unique across all service objects returned in this response. MUST be a non-empty string.
		Name:        normalizeDisplayName(service.DisplayName, string(service.ID)),
		Description: service.LongDescription,
		Bindable:    bindable,
		Metadata:    metadata,
		Plans:       []osb.Plan{plan},
		Tags:        service.Tags,
	}

	return osbService, nil
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

func normalizeDisplayName(name, id string) string {
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
