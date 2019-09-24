package populator

import (
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal"
)

const informerResyncPeriod = 30 * time.Minute

//go:generate mockery -name=instanceInserter -output=automock -outpkg=automock -case=underscore
type instanceInserter interface {
	Insert(i *internal.Instance) error
}

//go:generate mockery -name=instanceConverter -output=automock -outpkg=automock -case=underscore
type instanceConverter interface {
	MapServiceInstance(in *v1beta1.ServiceInstance) *internal.Instance
}
