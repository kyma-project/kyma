package populator

import (
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-connector/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-connector/application-broker/internal/broker"
)

const informerResyncPeriod = 30 * time.Minute

//go:generate mockery -name=instanceInserter -output=automock -outpkg=automock -case=underscore
type instanceInserter interface {
	Insert(i *internal.Instance) error
}

//go:generate mockery -name=operationInserter -output=automock -outpkg=automock -case=underscore
type operationInserter interface {
	Insert(io *internal.InstanceOperation) error
}

//go:generate mockery -name=brokerProcesses -output=automock -outpkg=automock -case=underscore
type brokerProcesses interface {
	ProvisionProcess(broker.RestoreProvisionRequest) error
	DeprovisionProcess(broker.DeprovisionProcessRequest)
	NewOperationID() (internal.OperationID, error)
}

//go:generate mockery -name=applicationServiceIDSelector -output=automock -outpkg=automock -case=underscore
type applicationServiceIDSelector interface {
	SelectApplicationServiceID(string, string) internal.ApplicationServiceID
}

//go:generate mockery -name=instanceConverter -output=automock -outpkg=automock -case=underscore
type instanceConverter interface {
	MapServiceInstance(in *v1beta1.ServiceInstance) *internal.Instance
}
