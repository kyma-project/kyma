package k8s

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/types"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/extractor"

	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
)

type DiscoveryInterface interface {
	discovery.DiscoveryInterface
}

//go:generate mockery -name=DiscoveryInterface -case=underscore -output=automock -outpkg=automock

type resourceService struct {
	client DiscoveryInterface
}

func newResourceService(client DiscoveryInterface) *resourceService {
	return &resourceService{
		client: client,
	}
}

func (svc *resourceService) Create(namespace string, resource types.Resource) (*types.Resource, error) {
	if namespace != resource.Namespace {
		return nil, apierror.NewInvalid(pretty.Resource, apierror.ErrorFieldAggregate{
			apierror.NewInvalidField("namespace", resource.Namespace, fmt.Sprintf("namespace of provided object does not match the namespace sent on the request (%s)", namespace)),
		})
	}

	pluralName, err := extractor.GetPluralNameFromKind(resource.Kind, resource.APIVersion, svc.client)
	if err != nil {
		return nil, errors.Wrap(err, "while getting resource's plural name")
	}

	result := svc.client.RESTClient().Post().
		AbsPath(svc.getAPIPath(resource.APIVersion)).
		Namespace(resource.Namespace).
		Resource(pluralName).
		Body(resource.Body).
		Do(context.Background())
	err = result.Error()
	if err != nil {
		return nil, errors.Wrap(err, "while creating resource")
	}

	body, err := result.Raw()
	if err != nil {
		return nil, errors.Wrap(err, "while extracting raw result")
	}

	return &types.Resource{
		APIVersion: resource.APIVersion,
		Name:       resource.Name,
		Namespace:  resource.Namespace,
		Kind:       resource.Kind,
		Body:       body,
	}, nil
}

func (svc *resourceService) getAPIPath(apiVersion string) string {
	if apiVersion == "v1" {
		return "/api/v1"
	}
	return fmt.Sprintf("/apis/%s", apiVersion)
}
