package apicontroller

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type apiResolver struct {
	apiSvc       apiSvc
	apiConverter *apiConverter
}

func newApiResolver(lister apiSvc) (*apiResolver, error) {
	if lister == nil {
		return nil, errors.New("Nil pointer for apiSvc")
	}

	return &apiResolver{
		apiSvc:       lister,
		apiConverter: &apiConverter{},
	}, nil
}

func (ar *apiResolver) APIsQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	apisObj, err := ar.apiSvc.List(namespace, serviceName, hostname)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for service name %v, hostname %v", pretty.APIs, serviceName, hostname))
		return nil, gqlerror.New(err, pretty.APIs, gqlerror.WithNamespace(namespace))
	}
	apis := ar.apiConverter.ToGQLs(apisObj)
	return apis, nil
}

func (ar *apiResolver) APIQuery(ctx context.Context, name string, namespace string) (*gqlschema.API, error) {
	apiObj, err := ar.apiSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return nil, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	if apiObj == nil {
		return nil, nil
	}

	api := ar.apiConverter.ToGQL(apiObj)
	return api, nil
}

func (ar *apiResolver) CreateAPI(ctx context.Context, name string, namespace string, params gqlschema.APIInput) (gqlschema.API, error) {

	api, err := ar.apiSvc.Create(name, namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return gqlschema.API{}, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return *ar.apiConverter.ToGQL(api), nil
}

func (ar *apiResolver) ApiEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan gqlschema.ApiEvent, error) {
	channel := make(chan gqlschema.ApiEvent, 1)
	filter := func(api *v1alpha2.Api) bool {
		if serviceName == nil {
			return api != nil && api.Namespace == namespace
		}
		return api != nil && api.Namespace == namespace && api.Spec.Service.Name == *serviceName
	}

	apiListener := listener.NewApi(channel, filter, ar.apiConverter)

	ar.apiSvc.Subscribe(apiListener)
	go func() {
		defer close(channel)
		defer ar.apiSvc.Unsubscribe(apiListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (ar *apiResolver) UpdateAPI(ctx context.Context, name string, namespace string, params gqlschema.APIInput) (gqlschema.API, error) {
	api, err := ar.apiSvc.Update(name, namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while editing %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return gqlschema.API{}, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return *ar.apiConverter.ToGQL(api), nil
}

func (ar *apiResolver) DeleteAPI(ctx context.Context, name string, namespace string) (*gqlschema.API, error) {
	apiObj, err := ar.apiSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return nil, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	apiCopy := apiObj.DeepCopy()
	err = ar.apiSvc.Delete(name, namespace)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.API, name, namespace))
		return nil, gqlerror.New(err, pretty.API, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	deletedAPI := ar.apiConverter.ToGQL(apiCopy)
	return deletedAPI, nil
}
