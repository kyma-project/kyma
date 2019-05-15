package application

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	"io/ioutil"
	"net/http"

	"crypto/tls"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	assetstorePretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/pretty"
	contentPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

const (
	kymaIntegrationNamespace = "kyma-integration"
)

type eventActivationResolver struct {
	service             eventActivationLister
	converter           *eventActivationConverter
	contentRetriever    shared.ContentRetriever
	assetStoreRetriever shared.AssetStoreRetriever
	verifySSL           bool
}

//go:generate mockery -name=eventActivationLister -output=automock -outpkg=automock -case=underscore
type eventActivationLister interface {
	List(namespace string) ([]*v1alpha1.EventActivation, error)
}

func newEventActivationResolver(service eventActivationLister, contentRetriever shared.ContentRetriever, assetStoreRetriever shared.AssetStoreRetriever, verifySSL bool) *eventActivationResolver {
	return &eventActivationResolver{
		service:             service,
		converter:           &eventActivationConverter{},
		contentRetriever:    contentRetriever,
		assetStoreRetriever: assetStoreRetriever,
		verifySSL:           verifySSL,
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

	asyncApiSpec, err := r.contentRetriever.AsyncApiSpec().Find("service-class", eventActivation.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.EventActivationEvents, pretty.EventActivation, eventActivation.Name))
		return nil, gqlerror.New(err, pretty.EventActivationEvents, gqlerror.WithName(eventActivation.Name))
	}

	if asyncApiSpec == nil {
		asyncApiSpec, err = r.getAsyncApi(eventActivation.Name)
		if err != nil {
			return nil, err
		}
	}

	if asyncApiSpec == nil {
		return []gqlschema.EventActivationEvent{}, nil
	}

	if asyncApiSpec.Data.AsyncAPI != "1.0.0" {
		details := fmt.Sprintf("not supported version `%s` of %s", asyncApiSpec.Data.AsyncAPI, contentPretty.AsyncApiSpec)
		glog.Error(details)
		return nil, gqlerror.NewInternal(gqlerror.WithDetails(details))
	}

	return r.converter.ToGQLEvents(asyncApiSpec), nil
}

// TODO: Remove this temporary function after removing content domain
func (r *eventActivationResolver) getAsyncApi(eventActivationName string) (*storage.AsyncApiSpec, error) {
	types := []string{"asyncapi", "asyncApi", "asyncapispec", "asyncApiSpec", "events"}

	items, err := r.assetStoreRetriever.ClusterAsset().ListForDocsTopicByType(eventActivationName, types)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", assetstorePretty.ClusterAssets, pretty.EventActivation, eventActivationName))
		return nil, gqlerror.New(err, assetstorePretty.ClusterAssets)
	}

	if len(items) == 0 {
		return nil, nil
	}

	assetRef := items[0].Status.AssetRef
	asyncApiFilePath := fmt.Sprintf("%s/%s", assetRef.BaseURL, assetRef.Files[0].Name)

	raw, err := r.fetchAsyncApi(asyncApiFilePath)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while fetching `AsyncApiSpec` for %s %s", pretty.EventActivation, eventActivationName))
		return nil, gqlerror.New(err, assetstorePretty.ClusterAsset)
	}

	asyncApi := new(storage.AsyncApiSpec)
	err = asyncApi.Decode(raw)
	if err != nil {
		return nil, err
	}

	return asyncApi, nil
}

// TODO: Remove this temporary function after removing content domain
func (r *eventActivationResolver) fetchAsyncApi(path string) ([]byte, error) {
	client := &http.Client{}
	if !r.verifySSL {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid SSL certificates
		}

		client.Transport = transCfg
	}

	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
