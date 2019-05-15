package application

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	assetstorePretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type eventActivationResolver struct {
	service             eventActivationLister
	converter           *eventActivationConverter
	assetStoreRetriever shared.AssetStoreRetriever
}

//go:generate mockery -name=eventActivationLister -output=automock -outpkg=automock -case=underscore
type eventActivationLister interface {
	List(namespace string) ([]*v1alpha1.EventActivation, error)
}

func newEventActivationResolver(service eventActivationLister, assetStoreRetriever shared.AssetStoreRetriever) *eventActivationResolver {
	return &eventActivationResolver{
		service:             service,
		converter:           &eventActivationConverter{},
		assetStoreRetriever: assetStoreRetriever,
	}
}

func (r *eventActivationResolver) EventActivationsQuery(ctx context.Context, namespace string) ([]gqlschema.EventActivation, error) {
	items, err := r.service.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s in `%s` namespace", pretty.EventActivations, namespace))
		return nil, gqlerror.New(err, pretty.EventActivations, gqlerror.WithNamespace(namespace))
	}

	return r.converter.ToGQLs(items), nil
}

func (r *eventActivationResolver) EventActivationEventsField(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	if eventActivation == nil {
		glog.Errorf("EventActivation cannot be empty in order to resolve events field")
		return nil, gqlerror.NewInternal()
	}

	types := []string{"asyncapi", "asyncApi", "asyncapispec", "asyncApiSpec", "events"}
	items, err := r.assetStoreRetriever.ClusterAsset().ListForDocsTopicByType(eventActivation.Name, types)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", assetstorePretty.ClusterAssets, pretty.EventActivation, eventActivation.Name))
		return nil, gqlerror.New(err, assetstorePretty.ClusterAssets)
	}

	if len(items) == 0 {
		return nil, nil
	}

	assetRef := items[0].Status.AssetRef
	asyncApiSpec, err := r.assetStoreRetriever.Specification().AsyncApi(assetRef.BaseURL, assetRef.Files[0].Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while fetching and decoding `AsyncApiSpec` for %s %s", pretty.EventActivation, eventActivation.Name))
		return []gqlschema.EventActivationEvent{}, gqlerror.New(err, assetstorePretty.ClusterAsset)
	}

	if asyncApiSpec.Data.AsyncAPI != "1.0.0" {
		details := fmt.Sprintf("not supported version `%s` of %s", asyncApiSpec.Data.AsyncAPI, "AsyncApiSpec")
		glog.Error(details)
		return nil, gqlerror.NewInternal(gqlerror.WithDetails(details))
	}

	return r.converter.ToGQLEvents(asyncApiSpec), nil
}
