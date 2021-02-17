package broker

import (
	"context"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/director"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	"github.com/kyma-project/kyma/components/application-broker/internal/servicecatalog"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	listers "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/platform/idprovider"

	gcli "github.com/kyma-project/kyma/components/application-broker/third_party/machinebox/graphql"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
	securityclientv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=instanceStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=appFinder -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=instanceGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=operationStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=removalProcessor -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=apiPackageCredentialsCreator -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=apiPackageCredentialsGetter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=apiPackageCredentialsRemover -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=ServiceBindingFetcher -output=automock -outpkg=automock -case=underscore

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

	apiPackageCredentialsCreator interface {
		EnsureAPIPackageCredentials(ctx context.Context, appID, pkgID, instanceID string, inputSchema map[string]interface{}) error
	}
	apiPackageCredentialsGetter interface {
		GetAPIPackageCredentials(ctx context.Context, appID, pkgID, instanceID string) (internal.APIPackageCredential, error)
	}
	apiPackageCredentialsRemover interface {
		EnsureAPIPackageCredentialsDeleted(ctx context.Context, appID string, pkgID string, instanceID string) error
	}

	DirectorService interface {
		apiPackageCredentialsCreator
		apiPackageCredentialsGetter
		apiPackageCredentialsRemover
	}

	ServiceBindingFetcher interface {
		GetServiceBindingSecretName(ns, externalID string) (string, error)
	}

	appSvcIDSelector interface {
		SelectID(req interface{}) internal.ApplicationServiceID
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
	istioClient *securityclientv1beta1.SecurityV1beta1Interface,
	log *logrus.Entry,
	livenessCheckStatus *LivenessCheckStatus,
	apiPackagesSupport bool,
	service director.ServiceConfig, directorProxyURL string,
	sbInformer cache.SharedIndexInformer, gatewayBaseURL string,
	idSelector appSvcIDSelector,
) *Server {

	idpRaw := idprovider.New()
	idp := func() (internal.OperationID, error) {
		idRaw, err := idpRaw()
		if err != nil {
			return "", err
		}
		return internal.OperationID(idRaw), nil
	}

	enabledChecker := access.NewApplicationMappingService(emLister)

	directorSvc, conv, getBindingCredentials, validateProvisionReq := getImplementationBasedOnVersion(sbInformer, service, directorProxyURL, gatewayBaseURL, apiPackagesSupport)

	stateService := &instanceStateService{operationCollectionGetter: opStorage}
	return &Server{
		catalogGetter: &catalogService{
			finder:            applicationFinder,
			conv:              conv,
			appEnabledChecker: enabledChecker,
		},
		provisioner: NewProvisioner(instStorage, stateService, opStorage, opStorage, accessChecker, applicationFinder,
			eaClient, knClient, *istioClient, instStorage, idp, log, idSelector, directorSvc, validateProvisionReq),
		deprovisioner: NewDeprovisioner(instStorage, stateService, opStorage, opStorage, idp, applicationFinder,
			knClient, eaClient, log, idSelector, directorSvc),
		binder: &bindService{
			appSvcFinder:     applicationFinder,
			appSvcIDSelector: idSelector,
			getCreds:         getBindingCredentials,
		},
		lastOpGetter: &getLastOperationService{
			getter: opStorage,
		},
		brokerService:       brokerService,
		sanityChecker:       NewSanityChecker(mClient, log, livenessCheckStatus),
		logger:              log.WithField("service", "broker:server"),
		operationIDProvider: idp,
	}
}

func getImplementationBasedOnVersion(sbInformer cache.SharedIndexInformer, service director.ServiceConfig, directorProxyURL string, gatewayBaseURL string, apiPackagesSupport bool) (DirectorService, converter, getCredentialFn, func(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError) {
	if apiPackagesSupport {
		sbFetcher := servicecatalog.NewServiceBindingFetcher(sbInformer)
		directorCli := director.NewQGLClient(gcli.NewClient(directorProxyURL))
		directorSvc := director.NewService(directorCli, service)
		credRenderer := NewBindingCredentialsRenderer(directorSvc, gatewayBaseURL, sbFetcher)

		return directorSvc, &appToServiceConverterV2{}, credRenderer.GetBindingCredentialsV2, validateProvisionRequestV2
	} else {
		directorSvc := director.NewNothingDoerService()
		credRenderer := BindingCredentialsRenderer{}
		return directorSvc, &appToServiceConverter{}, credRenderer.GetBindingCredentialsV1, validateProvisionRequestV1
	}
}
