package broker

import (
	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"

	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/platform/idprovider"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybind"
)

// be aware that after regenerating mocks, manual steps are required
//go:generate mockery -name=bundleStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=chartGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=chartStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=operationStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=helmClient -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceStateGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceBindDataGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceBindDataRemover -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceBindDataInserter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=bindTemplateRenderer -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=bindTemplateResolver -output=automock -outpkg=automock -case=underscore

type (
	bundleIDGetter interface {
		GetByID(id internal.BundleID) (*internal.Bundle, error)
	}
	bundleFinder interface {
		FindAll() ([]*internal.Bundle, error)
	}
	bundleStorage interface {
		bundleIDGetter
		bundleFinder
	}

	chartGetter interface {
		Get(name internal.ChartName, ver semver.Version) (*chart.Chart, error)
	}
	chartStorage interface {
		chartGetter
	}

	operationInserter interface {
		Insert(io *internal.InstanceOperation) error
	}
	operationGetter interface {
		Get(iID internal.InstanceID, opID internal.OperationID) (*internal.InstanceOperation, error)
	}
	operationCollectionGetter interface {
		GetAll(iID internal.InstanceID) ([]*internal.InstanceOperation, error)
	}
	operationUpdater interface {
		UpdateState(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState) error
		UpdateStateDesc(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState, desc *string) error
	}
	operationRemover interface {
		Remove(iID internal.InstanceID, opID internal.OperationID) error
	}
	operationStorage interface {
		operationInserter
		operationGetter
		operationCollectionGetter
		operationUpdater
		operationRemover
	}

	instanceInserter interface {
		Insert(i *internal.Instance) error
	}
	instanceGetter interface {
		Get(id internal.InstanceID) (*internal.Instance, error)
	}
	instanceRemover interface {
		Remove(id internal.InstanceID) error
	}
	instanceStorage interface {
		instanceInserter
		instanceGetter
		instanceRemover
	}

	instanceStateProvisionGetter interface {
		IsProvisioned(internal.InstanceID) (bool, error)
		IsProvisioningInProgress(internal.InstanceID) (internal.OperationID, bool, error)
	}

	instanceStateDeprovisionGetter interface {
		IsDeprovisioned(internal.InstanceID) (bool, error)
		IsDeprovisioningInProgress(internal.InstanceID) (internal.OperationID, bool, error)
	}

	instanceStateGetter interface {
		instanceStateProvisionGetter
		instanceStateDeprovisionGetter
	}

	helmInstaller interface {
		Install(c *chart.Chart, cv internal.ChartValues, releaseName internal.ReleaseName, namespace internal.Namespace) (*rls.InstallReleaseResponse, error)
	}
	helmDeleter interface {
		Delete(internal.ReleaseName) error
	}
	helmClient interface {
		helmInstaller
		helmDeleter
	}

	instanceBindDataGetter interface {
		Get(iID internal.InstanceID) (*internal.InstanceBindData, error)
	}

	instanceBindDataInserter interface {
		Insert(*internal.InstanceBindData) error
	}

	instanceBindDataRemover interface {
		Remove(internal.InstanceID) error
	}

	instanceBindDataStorage interface {
		instanceBindDataGetter
		instanceBindDataInserter
		instanceBindDataRemover
	}

	bindTemplateRenderer interface {
		Render(bindTemplate internal.BundlePlanBindTemplate, resp *rls.InstallReleaseResponse) (ybind.RenderedBindYAML, error)
	}

	bindTemplateResolver interface {
		Resolve(bindYAML ybind.RenderedBindYAML, ns internal.Namespace) (*ybind.ResolveOutput, error)
	}
)

// New creates instance of broker.
func New(bs bundleStorage, cs chartStorage, os operationStorage, is instanceStorage, ibd instanceBindDataStorage,
	bindTmplRenderer bindTemplateRenderer, bindTmplResolver bindTemplateResolver, hc helmClient, log *logrus.Entry) *Server {
	idpRaw := idprovider.New()
	idp := func() (internal.OperationID, error) {
		idRaw, err := idpRaw()
		if err != nil {
			return internal.OperationID(""), err
		}
		return internal.OperationID(idRaw), nil
	}

	return newWithIDProvider(bs, cs, os, is, ibd, bindTmplRenderer, bindTmplResolver, hc, log, idp)
}

func newWithIDProvider(bs bundleStorage, cs chartStorage, os operationStorage, is instanceStorage, ibd instanceBindDataStorage,
	bindTmplRenderer bindTemplateRenderer, bindTmplResolver bindTemplateResolver, hc helmClient,
	log *logrus.Entry, idp func() (internal.OperationID, error)) *Server {
	return &Server{
		catalogGetter: &catalogService{
			finder: bs,
			conv:   &bundleToServiceConverter{},
		},
		provisioner: &provisionService{
			bundleIDGetter:   bs,
			chartGetter:      cs,
			instanceInserter: is,
			instanceStateGetter: &instanceStateService{
				operationCollectionGetter: os,
			},
			operationInserter:        os,
			operationUpdater:         os,
			operationIDProvider:      idp,
			helmInstaller:            hc,
			log:                      log.WithField("service", "provisioner"),
			bindTemplateRenderer:     bindTmplRenderer,
			bindTemplateResolver:     bindTmplResolver,
			instanceBindDataInserter: ibd,
		},
		deprovisioner: &deprovisionService{
			instanceGetter:    is,
			operationInserter: os,
			instanceStateGetter: &instanceStateService{
				operationCollectionGetter: os,
			},
			operationUpdater:        os,
			instanceBindDataRemover: ibd,
			operationIDProvider:     idp,
			helmDeleter:             hc,
		},
		binder: &bindService{
			instanceBindDataGetter: ibd,
		},
		unbinder: &unbindService{},
		lastOpGetter: &getLastOperationService{
			getter: os,
		},
		logger: log.WithField("service", "server"),
	}
}
