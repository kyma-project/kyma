package application

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	rafterPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type eventActivationResolver struct {
	service         eventActivationLister
	converter       *eventActivationConverter
	rafterRetriever shared.RafterRetriever
}

//go:generate mockery -name=eventActivationLister -output=automock -outpkg=automock -case=underscore
type eventActivationLister interface {
	List(namespace string) ([]*v1alpha1.EventActivation, error)
}

func newEventActivationResolver(service eventActivationLister, rafterRetriever shared.RafterRetriever) *eventActivationResolver {
	return &eventActivationResolver{
		service:         service,
		converter:       &eventActivationConverter{},
		rafterRetriever: rafterRetriever,
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

	types := []string{"asyncapi", "asyncApi", "asyncapispec", "asyncApiSpec", "events", "async-api"}
	items, err := r.rafterRetriever.ClusterAsset().ListForClusterAssetGroupByType(eventActivation.Name, types)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", rafterPretty.ClusterAssets, pretty.EventActivation, eventActivation.Name))
		return nil, gqlerror.New(err, rafterPretty.ClusterAssets)
	}

	if len(items) == 0 || items[0].Status.Phase != v1beta1.AssetReady || len(items[0].Status.AssetRef.Files) == 0 {
		return nil, nil
	}

	assetRef := items[0].Status.AssetRef
	asyncApiSpec, err := r.rafterRetriever.Specification().AsyncAPI(assetRef.BaseURL, assetRef.Files[0].Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while fetching and decoding `AsyncApiSpec` for %s %s", pretty.EventActivation, eventActivation.Name))
		return []gqlschema.EventActivationEvent{}, gqlerror.New(err, rafterPretty.ClusterAsset)
	}

	if asyncApiSpec.Data.AsyncAPI != "2.0.0" {
		details := fmt.Sprintf("not supported version `%s` of %s", asyncApiSpec.Data.AsyncAPI, "AsyncApiSpec")
		glog.Error(details)
		return nil, gqlerror.NewInternal(gqlerror.WithDetails(details))
	}

	return r.converter.ToGQLEvents(asyncApiSpec), nil
}
