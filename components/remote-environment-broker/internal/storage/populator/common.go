package populator

import (
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
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
