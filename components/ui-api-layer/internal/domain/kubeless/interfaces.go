package kubeless

import (
	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
)

//go:generate mockery -name=functionLister -output=automock -outpkg=automock -case=underscore
type functionLister interface {
	List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.Function, error)
}
