package serverless

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/convert"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

func (r *resolver) FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error) {
	items, err := r.functionService.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Functions))
		return nil, gqlerror.New(err, pretty.Functions)
	}

	functions, err := convert.FunctionsToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Functions))
		return nil, gqlerror.New(err, pretty.Functions)
	}

	return functions, nil
}
