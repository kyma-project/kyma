package broker

import (
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

func NewWithIDProvider(bs bundleStorage, cs chartStorage, os operationStorage, is instanceStorage, ibd instanceBindDataStorage,
	bindTmplRenderer bindTemplateRenderer, bindTmplResolver bindTemplateResolver,
	hc helmClient, log *logrus.Entry, idp func() (internal.OperationID, error)) *Server {
	return newWithIDProvider(bs, cs, os, is, ibd, bindTmplRenderer, bindTmplResolver, hc, log, idp)
}
