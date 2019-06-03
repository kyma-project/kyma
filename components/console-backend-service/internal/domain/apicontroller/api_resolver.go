package apicontroller

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type apiResolver struct {
	apiLister    apiLister
	apiConverter apiConverter
}

func newApiResolver(lister apiLister) (*apiResolver, error) {
	if lister == nil {
		return nil, errors.New("Nil pointer for apiLister")
	}

	return &apiResolver{
		apiLister:    lister,
		apiConverter: apiConverter{},
	}, nil
}

func (ar *apiResolver) APIsQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	apis, err := ar.apiLister.List(namespace, serviceName, hostname)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for service name %v, hostname %v", pretty.APIs, serviceName, hostname))
		return nil, gqlerror.New(err, pretty.APIs, gqlerror.WithNamespace(namespace))
	}

	return ar.apiConverter.ToGQLs(apis), nil
}


func (ar *apiResolver) CreateAPI(ctx context.Context, name string, namespace string, hostname string, serviceName string, servicePort int, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (gqlschema.API, error) {
	api, err := ar.apiLister.Create(name, namespace, hostname, serviceName, servicePort, disableIstioAuthPolicyMTLS, authenticationEnabled)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s %v", pretty.APIs, name))
		return gqlschema.API{}, gqlerror.New(err, pretty.APIs, gqlerror.WithNamespace(namespace))
	}

	return gqlschema.API{
		Name: api.Name,
		Hostname: api.Spec.Hostname,
		Service: gqlschema.ApiService{
			Name: api.Spec.Service.Name,
			Port: api.Spec.Service.Port,
		},
	}, nil
	// AuthenticationPolicies
}
