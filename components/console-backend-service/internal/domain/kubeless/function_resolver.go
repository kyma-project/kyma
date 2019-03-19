package kubeless

import (
	"context"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/kubeless/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
)

//go:generate mockery -name=functionLister -output=automock -outpkg=automock -case=underscore
type functionLister interface {
	List(namespace string, pagingParams pager.PagingParams) ([]*v1beta1.Function, error)
}

type functionResolver struct {
	functionLister    functionLister
	functionConverter *functionConverter
}

func newFunctionResolver(functionLister functionLister) (*functionResolver, error) {
	if functionLister == nil {
		return nil, errors.New("Nil pointer for functionLister")
	}

	return &functionResolver{
		functionLister:    functionLister,
		functionConverter: &functionConverter{},
	}, nil
}

func (r *functionResolver) FunctionsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Function, error) {
	functions, err := r.functionLister.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for namespace %s", pretty.Functions, namespace))
		return nil, gqlerror.New(err, pretty.Functions, gqlerror.WithNamespace(namespace))
	}

	return r.functionConverter.ToGQLs(functions), nil
}
