package apicontroller

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type apiResolver struct {
	apiLister    apiLister
	apiConverter apiConverter
}

func newApiResolver(lister apiLister) *apiResolver {
	return &apiResolver{
		apiLister:    lister,
		apiConverter: apiConverter{},
	}
}

func (ar *apiResolver) APIsQuery(ctx context.Context, environment string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	apis, err := ar.apiLister.List(environment, serviceName, hostname)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing APIs for service name %s, hostname %s", serviceName, hostname))
		return nil, fmt.Errorf("cannot query APIs")
	}

	return ar.apiConverter.ToGQLs(apis), nil
}
