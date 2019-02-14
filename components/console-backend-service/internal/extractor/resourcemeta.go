package extractor

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/apierror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
)

type ResourceMeta struct {
	Name       string
	Namespace  string
	Kind       string
	APIVersion string
}

func ExtractResourceMeta(in map[string]interface{}) (ResourceMeta, error) {
	var errs apierror.ErrorFieldAggregate
	apiVersion, ok := in["apiVersion"].(string)
	if !ok {
		errs = append(errs, apierror.NewMissingField("apiVersion"))
	}
	kind, ok := in["kind"].(string)
	if !ok {
		errs = append(errs, apierror.NewMissingField("kind"))
	}
	metadata, ok := in["metadata"].(map[string]interface{})
	var name, namespace string
	if ok {
		name, ok = metadata["name"].(string)
		if !ok {
			errs = append(errs, apierror.NewMissingField("metadata.name"))
		}
		namespace, ok = metadata["namespace"].(string)
		if !ok {
			errs = append(errs, apierror.NewMissingField("metadata.namespace"))
		}
	} else {
		errs = append(errs, apierror.NewMissingField("metadata"))
	}
	if len(errs) > 0 {
		return ResourceMeta{}, apierror.NewInvalid(pretty.Resource, errs)
	}

	return ResourceMeta{
		Name:       name,
		Namespace:  namespace,
		Kind:       kind,
		APIVersion: apiVersion,
	}, nil
}

func GetPluralNameFromKind(kind, apiVersion string, client discovery.DiscoveryInterface) (string, error) {
	resources, err := client.ServerResourcesForGroupVersion(apiVersion)
	if err != nil {
		return "", errors.Wrapf(err, "while fetching resources for group version %s", apiVersion)
	}

	var plural string
	for _, resource := range resources.APIResources {
		if resource.Kind == kind {
			plural = resource.Name
			break
		}
	}

	if plural == "" {
		return "", apierror.NewInvalid(pretty.Resource, apierror.ErrorFieldAggregate{apierror.NewInvalidField("kind", kind, "resource plural name for specified kind not found")})
	}

	return plural, nil
}
