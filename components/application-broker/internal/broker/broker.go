package broker

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	listers "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/platform/idprovider"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
	istioCli "istio.io/client-go/pkg/clientset/versioned"
)

//go:generate mockery -name=instanceStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=appFinder -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=operationStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=removalProcessor -output=automock -outpkg=automock -case=underscore

type (
	applicationFinder interface {
		FindAll() ([]*internal.Application, error)
		Get(name internal.ApplicationName) (*internal.Application, error)
	}
	appSvcFinder interface {
		FindOneByServiceID(id internal.ApplicationServiceID) (*internal.Application, error)
	}
	appFinder interface {
		applicationFinder
		appSvcFinder
	}
	operationInserter interface {
		Insert(io *internal.InstanceOperation) error
	}
	operationGetter interface {
		Get(iID internal.InstanceID, opID internal.OperationID) (*internal.InstanceOperation, error)
	}
	operationCollectionGetter interface {
		GetAll(iID internal.InstanceID) ([]*internal.InstanceOperation, error)
		GetLast(iID internal.InstanceID) (*internal.InstanceOperation, error)
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
	instanceFinder interface {
		FindOne(m func(i *internal.Instance) bool) (*internal.Instance, error)
		FindAll(m func(i *internal.Instance) bool) ([]*internal.Instance, error)
	}
	instanceStateUpdater interface {
		UpdateState(iID internal.InstanceID, state internal.InstanceState) error
	}
	instanceStorage interface {
		instanceInserter
		instanceGetter
		instanceRemover
		instanceFinder
		instanceStateUpdater
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
)

// New creates instance of broker server.
func New(applicationFinder appFinder,
	instStorage instanceStorage,
	opStorage operationStorage,
	accessChecker access.ProvisionChecker,
	eaClient v1alpha1.ApplicationconnectorV1alpha1Interface,
	emLister listers.ApplicationMappingLister,
	brokerService *NsBrokerService,
	mClient *mappingCli.Interface,
	knClient knative.Client,
	istioClient *istioCli.Interface,
	log *logrus.Entry,
	livenessCheckStatus *LivenessCheckStatus,
	apiPackagesSupport bool,
) *Server {

	idpRaw := idprovider.New()
	idp := func() (internal.OperationID, error) {
		idRaw, err := idpRaw()
		if err != nil {
			return internal.OperationID(""), err
		}
		return internal.OperationID(idRaw), nil
	}

	enabledChecker := access.NewApplicationMappingService(emLister)

	conv, getBindingCredentials, idSelector := getImplementationBasedOnVersion(apiPackagesSupport)

	stateService := &instanceStateService{operationCollectionGetter: opStorage}
	return &Server{
		catalogGetter: &catalogService{
			finder:            applicationFinder,
			conv:              conv,
			appEnabledChecker: enabledChecker,
		},
		provisioner: NewProvisioner(instStorage, instStorage, stateService, opStorage, opStorage, accessChecker,
			applicationFinder, eaClient, knClient, *istioClient, instStorage, idp, log, idSelector),
		deprovisioner: NewDeprovisioner(instStorage, stateService, opStorage, opStorage, idp, applicationFinder,
			knClient, eaClient, log, idSelector),
		binder: &bindService{
			appSvcFinder:     applicationFinder,
			appSvcIDSelector: idSelector,
			getCreds:         getBindingCredentials,
		},
		lastOpGetter: &getLastOperationService{
			getter: opStorage,
		},
		brokerService: brokerService,
		sanityChecker: NewSanityChecker(mClient, log, livenessCheckStatus),
		logger:        log.WithField("service", "broker:server"),
	}
}

func getImplementationBasedOnVersion(apiPackagesSupport bool) (converter, func(entries []internal.Entry) map[string]interface{}, *IDSelector) {
	if apiPackagesSupport {
		return &appToServiceConverterV2{}, getBindingCredentialsV2, &IDSelector{apiPackagesSupport: apiPackagesSupport}
	} else {
		return &appToServiceConverter{}, getBindingCredentialsV1, &IDSelector{apiPackagesSupport: apiPackagesSupport}
	}
}

type AppSvcIDSelector interface {
	SelectID(req interface{}) internal.ApplicationServiceID
}

type IDSelector struct {
	apiPackagesSupport bool
}

func (s *IDSelector) SelectID(req interface{}) internal.ApplicationServiceID {
	var svcID, planID string
	switch d := req.(type) {
	case *osb.BindRequest:
		svcID, planID = d.ServiceID, d.PlanID
	case *osb.ProvisionRequest:
		svcID, planID = d.ServiceID, d.PlanID
	case *osb.DeprovisionRequest:
		svcID, planID = d.ServiceID, d.PlanID
	}

	// In new approach ApplicationServiceID == req Plan ID
	if s.apiPackagesSupport {
		return internal.ApplicationServiceID(planID)
	}

	// In old approach ApplicationServiceID == req Service ID == Class ID
	return internal.ApplicationServiceID(svcID)
}
