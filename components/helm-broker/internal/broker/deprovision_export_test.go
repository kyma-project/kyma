package broker

import "github.com/kyma-project/kyma/components/helm-broker/internal"

func NewDeprovisionService(is instanceStorage, oi operationInserter, ou operationUpdater, ibdr instanceBindDataRemover, hd helmDeleter, oIDProv func() (internal.OperationID, error), isg instanceStateGetter) *deprovisionService {
	return &deprovisionService{
		instanceGetter:          is,
		instanceRemover:         is,
		instanceStateGetter:     isg,
		operationInserter:       oi,
		operationUpdater:        ou,
		operationIDProvider:     oIDProv,
		instanceBindDataRemover: ibdr,
		helmDeleter:             hd,
	}
}

func (svc *deprovisionService) WithTestHookOnAsyncCalled(h func(internal.OperationID)) *deprovisionService {
	svc.testHookAsyncCalled = h
	return svc
}
