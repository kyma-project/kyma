package broker

import (
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

func NewProvisionService(bg bundleIDGetter, cg chartGetter, is instanceStorage, isg instanceStateGetter, oi operationInserter, ou operationUpdater,
	ibd instanceBindDataInserter, bindTmplRenderer bindTemplateRenderer, bindTmplResolver bindTemplateResolver,
	hi helmInstaller, oIDProv func() (internal.OperationID, error), log *logrus.Entry) *provisionService {
	return &provisionService{
		bundleIDGetter:           bg,
		chartGetter:              cg,
		instanceGetter:           is,
		instanceInserter:         is,
		instanceStateGetter:      isg,
		operationInserter:        oi,
		operationUpdater:         ou,
		operationIDProvider:      oIDProv,
		helmInstaller:            hi,
		log:                      log,
		instanceBindDataInserter: ibd,
		bindTemplateRenderer:     bindTmplRenderer,
		bindTemplateResolver:     bindTmplResolver,
	}
}

func (svc *provisionService) WithTestHookOnAsyncCalled(h func(internal.OperationID)) *provisionService {
	svc.testHookAsyncCalled = h
	return svc
}
