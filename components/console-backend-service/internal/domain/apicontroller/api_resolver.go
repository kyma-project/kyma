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
	apiLister    apiLister
	apiConverter *apiConverter
}

func newApiResolver(lister apiLister) (*apiResolver, error) {
	if lister == nil {
		return nil, errors.New("Nil pointer for apiLister")
	}

	return &apiResolver{
		apiLister:    lister,
		apiConverter: &apiConverter{},
	}, nil
}

func (ar *apiResolver) APIsQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	apisObj, err := ar.apiLister.List(namespace, serviceName, hostname)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for service name %v, hostname %v", pretty.APIs, serviceName, hostname))
		return nil, gqlerror.New(err, pretty.APIs, gqlerror.WithNamespace(namespace))
	}
	apis := ar.apiConverter.ToGQLs(apisObj)
	return apis, nil
}

func (ar *apiResolver) APIQuery(ctx context.Context, name string, namespace string) (*gqlschema.API, error) {
	apiObj, err := ar.apiLister.Find(name, namespace)
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

func (ar *apiResolver) CreateAPI(ctx context.Context, name string, namespace string, hostname string, serviceName string, servicePort int, jwksUri string, issuer string, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (gqlschema.API, error) {
	api, err := ar.apiLister.Create(name, namespace, hostname, serviceName, servicePort, jwksUri, issuer, disableIstioAuthPolicyMTLS, authenticationEnabled)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return gqlschema.API{}, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return gqlschema.API{
		Name:     api.Name,
		Hostname: api.Spec.Hostname,
		Service: gqlschema.ApiService{
			Name: api.Spec.Service.Name,
			Port: api.Spec.Service.Port,
		},
		AuthenticationPolicies: []gqlschema.AuthenticationPolicy{
			{
				JwksURI: jwksUri,
				Issuer:  issuer,
				Type:    gqlschema.AuthenticationPolicyType("JWT"),
			},
		},
	}, nil
}

func (ar *apiResolver) ApiEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ApiEvent, error) {
	channel := make(chan gqlschema.ApiEvent, 1)
	filter := func(api *v1alpha2.Api) bool {
		return api != nil && api.Namespace == namespace
	}

	apiListener := listener.NewApi(channel, filter, ar.apiConverter)

	ar.apiLister.Subscribe(apiListener)
	go func() {
		defer close(channel)
		defer ar.apiLister.Unsubscribe(apiListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (ar *apiResolver) UpdateAPI(ctx context.Context, name string, namespace string, hostname string, serviceName string, servicePort int, jwksUri string, issuer string, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (gqlschema.API, error) {
	api, err := ar.apiLister.Update(name, namespace, hostname, serviceName, servicePort, jwksUri, issuer, disableIstioAuthPolicyMTLS, authenticationEnabled)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while editing %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return gqlschema.API{}, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return gqlschema.API{
		Name:     api.Name,
		Hostname: api.Spec.Hostname,
		Service: gqlschema.ApiService{
			Name: api.Spec.Service.Name,
			Port: api.Spec.Service.Port,
		},
		AuthenticationPolicies: []gqlschema.AuthenticationPolicy{
			{
				JwksURI: jwksUri,
				Issuer:  issuer,
				Type:    gqlschema.AuthenticationPolicyType("JWT"),
			},
		},
	}, nil
}

func (ar *apiResolver) DeleteAPI(ctx context.Context, name string, namespace string) (*gqlschema.API, error) {
	apiObj, err := ar.apiLister.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.API, name, namespace))
		return nil, gqlerror.New(err, pretty.APIs, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	apiCopy := apiObj.DeepCopy()
	err = ar.apiLister.Delete(name, namespace)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.API, name, namespace))
		return nil, gqlerror.New(err, pretty.API, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	deletedAPI := ar.apiConverter.ToGQL(apiCopy)
	return deletedAPI, nil
}
