package broker

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	listers "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/platform/idprovider"

	istioversionedclient "istio.io/client-go/pkg/clientset/versioned"
)

//go:generate mockery -name=instanceStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=appFinder -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=serviceInstanceGetter -output=automock -outpkg=automock -case=underscore
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

	serviceInstanceGetter interface {
		GetByNamespaceAndExternalID(namespace string, extID string) (*v1beta1.ServiceInstance, error)
	}
)

// New creates instance of broker server.
func New(applicationFinder appFinder,
	instStorage instanceStorage,
	opStorage operationStorage,
	accessChecker access.ProvisionChecker,
	eaClient v1alpha1.ApplicationconnectorV1alpha1Interface,
	serviceInstanceGetter serviceInstanceGetter,
	emLister listers.ApplicationMappingLister,
	brokerService *NsBrokerService,
	mClient *mappingCli.Interface,
	knClient knative.Client,
	istioClient istioversionedclient.Interface,
	log *logrus.Entry,
	livenessCheckStatus *LivenessCheckStatus,
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

	stateService := &instanceStateService{operationCollectionGetter: opStorage}
	return &Server{
		catalogGetter: &catalogService{
			finder:            applicationFinder,
			conv:              &appToServiceConverter{},
			appEnabledChecker: enabledChecker,
		},
		provisioner:   NewProvisioner(instStorage, instStorage, stateService, opStorage, opStorage, accessChecker,
			applicationFinder, serviceInstanceGetter, eaClient, knClient, istioClient, instStorage, idp, log),
		deprovisioner: NewDeprovisioner(instStorage, stateService, opStorage, opStorage, idp, applicationFinder,
			knClient, log),
		binder: &bindService{
			appSvcFinder: applicationFinder,
		},
		lastOpGetter: &getLastOperationService{
			getter: opStorage,
		},
		brokerService: brokerService,
		sanityChecker: NewSanityChecker(mClient, log, livenessCheckStatus),
		logger:        log.WithField("service", "broker:server"),
	}
}
