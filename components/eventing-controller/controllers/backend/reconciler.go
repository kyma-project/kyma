package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/internal/featureflags"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

//nolint:gosec
const (
	BEBBackendSecretLabelKey   = "kyma-project.io/eventing-backend"
	BEBBackendSecretLabelValue = "beb"

	BEBSecretNameSuffix = "-beb-oauth2"

	BackendCRLabelKey   = "kyma-project.io/eventing"
	BackendCRLabelValue = "backend"

	AppLabelValue             = deployment.PublisherName
	PublisherSecretEMSHostKey = "ems-publish-host"

	TokenEndpointFormat             = "%s?grant_type=%s&response_type=token"
	NamespacePrefix                 = "/"
	BEBPublishEndpointForSubscriber = "/sap/ems/v1"
	BEBPublishEndpointForPublisher  = "/sap/ems/v1/events"

	reconcilerName = "backend-reconciler"

	kymaSystemNamespace = "kyma-system"
	tlsCertField        = "tls.crt"

	natsSecretName           = "eventing-nats-secret"
	natsSecretKey            = "resolver.conf"
	natsSecretPasswordLength = 60

	secretKeyClientID     = "client_id"
	secretKeyClientSecret = "client_secret"
	secretKeyTokenURL     = "token_url"
	secretKeyCertsURL     = "certs_url"
)

var (
	// allowedAnnotations are the publisher proxy deployment spec template annotations
	// which should be preserved during reconciliation.
	allowedAnnotations = map[string]string{
		"kubectl.kubernetes.io/restartedAt": "",
	}

	errObjectNotFound = errors.New("object not found")
	errInvalidObject  = errors.New("invalid object")
)

type oauth2Credentials struct {
	clientID     []byte
	clientSecret []byte
	tokenURL     []byte
	certsURL     []byte
}

type Reconciler struct {
	client.Client
	ctx               context.Context
	natsSubMgr        subscriptionmanager.Manager
	natsConfig        env.NATSConfig
	natsSubMgrStarted bool
	bebSubMgr         subscriptionmanager.Manager
	bebSubMgrStarted  bool
	logger            *logger.Logger
	record            record.EventRecorder
	cfg               env.BackendConfig
	envCfg            env.Config
	// backendType is the type of the backend which the reconciler detects at runtime
	backendType eventingv1alpha1.BackendType
	// credentials that are passed to the BEB subscription reconciler
	credentials oauth2Credentials
}

func NewReconciler(
	ctx context.Context,
	natsSubMgr subscriptionmanager.Manager,
	natsConfig env.NATSConfig,
	envCfg env.Config,
	backendCfg env.BackendConfig,
	bebSubMgr subscriptionmanager.Manager,
	client client.Client,
	logger *logger.Logger,
	recorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		ctx:        ctx,
		natsSubMgr: natsSubMgr,
		natsConfig: natsConfig,
		envCfg:     envCfg,
		bebSubMgr:  bebSubMgr,
		Client:     client,
		logger:     logger,
		record:     recorder,
		cfg:        backendCfg,
	}
}

func (r *Reconciler) SetNatsConfig(natsConfig env.NATSConfig) {
	r.natsConfig = natsConfig
}

func (r *Reconciler) SetBackendConfig(backendCfg env.BackendConfig) {
	r.cfg = backendCfg
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete

func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	if err := r.updateMutatingValidatingWebhookWithCABundle(ctx); err != nil {
		return ctrl.Result{}, err
	}

	// Create NATS Secret for the eventing-nats statefulset
	if createErr := r.createNATSSecret(ctx); createErr != nil {
		return ctrl.Result{}, createErr
	}

	var secretList v1.SecretList
	// the default status has all conditions and eventingReady set to true.
	// if something breaks during reconciliation, the condition and eventingReady is updated to false.
	defaultStatus := getDefaultBackendStatus()

	if err := r.List(ctx, &secretList, client.MatchingLabels{
		BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
	}); err != nil {
		return ctrl.Result{}, err
	}

	if len(secretList.Items) > 1 {
		// This is not allowed!
		r.namedLogger().Debugw("More than one secret with the EventingBackend label exist", "key", BEBBackendSecretLabelKey, "value", BEBBackendSecretLabelValue, "count", len(secretList.Items))
		defaultStatus.Backend = eventingv1alpha1.BEBBackendType
		defaultStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionDuplicateSecrets, "")
		if updateErr := r.syncBackendStatus(ctx, &defaultStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(updateErr, "update EventingBackend status failed")
		}
		return ctrl.Result{}, nil
	}

	// If secret with label then BEB flow
	if len(secretList.Items) == 1 {
		return r.reconcileBEBBackend(ctx, &secretList.Items[0], &defaultStatus)
	}

	// Default: NATS flow
	return r.reconcileNATSBackend(ctx, &defaultStatus)
}

func (r *Reconciler) reconcileNATSBackend(ctx context.Context, backendStatus *eventingv1alpha1.EventingBackendStatus) (ctrl.Result, error) {
	r.backendType = eventingv1alpha1.NatsBackendType
	backendStatus.Backend = r.backendType
	// CreateOrUpdate CR with NATS
	err := r.CreateOrUpdateBackendCR(ctx)
	if err != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonBackendCRSyncFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while creating/updating of EventingBackend")
		}
		return ctrl.Result{}, errors.Wrapf(err, "create or update EventingBackend failed, type: %s", eventingv1alpha1.NatsBackendType)
	}

	// Stop the BEB subscription controller
	if err := r.stopBEBController(); err != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonControllerStopFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while stopping BEB controller")
		}
		return ctrl.Result{}, err
	}

	// Start the NATS subscription controller
	if err := r.startNATSController(); err != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonControllerStartFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while starting NATS controller")
		}
		return ctrl.Result{}, err
	}

	// Delete secret for publisher proxy if it exists
	err = r.deletePublisherProxySecret(ctx)
	if err != nil {
		backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonPublisherProxySecretError, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while deleting Event Publisher secret")
		}
		return ctrl.Result{}, errors.Wrapf(err, "delete eventing Event Publisher secret failed")
	}

	// CreateOrUpdate deployment for publisher proxy
	publisher, err := r.CreateOrUpdatePublisherProxy(ctx, r.backendType)
	if err != nil {
		backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonPublisherProxySyncFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while creating/updating Event Publisher deployment")
		}
		return ctrl.Result{}, err
	}

	if r.natsSubMgrStarted && !backendStatus.IsSubscriptionControllerStatusReady() {
		backendStatus.SetSubscriptionControllerReadyCondition(true, eventingv1alpha1.ConditionReasonSubscriptionControllerReady, "")
	}
	// CreateOrUpdate status of the CR
	// Get publisher proxy ready status
	err = r.syncBackendStatus(ctx, backendStatus, publisher)
	return ctrl.Result{}, err
}

func (r *Reconciler) reconcileBEBBackend(ctx context.Context, bebSecret *v1.Secret, backendStatus *eventingv1alpha1.EventingBackendStatus) (ctrl.Result, error) {
	r.backendType = eventingv1alpha1.BEBBackendType
	backendStatus.Backend, backendStatus.BEBSecretName, backendStatus.BEBSecretNamespace = r.backendType, bebSecret.Name, bebSecret.Namespace

	// CreateOrUpdate CR with BEB
	err := r.CreateOrUpdateBackendCR(ctx)
	if err != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonBackendCRSyncFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while creating/updating EventingBackend")
		}
		return ctrl.Result{}, errors.Wrapf(err, "create/update EventingBackend failed, type: %s", eventingv1alpha1.BEBBackendType)
	}

	// Stop the NATS subscription controller
	if err := r.stopNATSController(); err != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonControllerStopFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while stopping NATS controller")
		}
		return ctrl.Result{}, err
	}

	// gets oauth2ClientID and secret and stops the BEB controller if changed
	err = r.syncOauth2ClientIDAndSecret(ctx, backendStatus)
	if err != nil {
		backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonOauth2ClientSyncFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while syncing oauth2Client")
		}
		return ctrl.Result{}, err
	}

	// CreateOrUpdate deployment for publisher proxy secret
	secretForPublisher, err := r.SyncPublisherProxySecret(ctx, bebSecret)
	if err != nil {
		backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonPublisherProxySecretError, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while syncing Event Publisher secret")
		}
		return ctrl.Result{}, err
	}

	// Set environment with secrets for BEB subscription controller
	err = setUpEnvironmentForBEBController(secretForPublisher)
	if err != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonControllerStartFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while setting up environment variables for BEB controller")
		}
		return ctrl.Result{}, errors.Wrapf(err, "failed to setup environment variables for BEB controller")
	}

	// Start the BEB subscription controller
	if startErr := r.startBEBController(); startErr != nil {
		backendStatus.SetSubscriptionControllerReadyCondition(false,
			eventingv1alpha1.ConditionReasonControllerStartFailed, startErr.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(startErr, "failed to update status while starting BEB controller")
		}
		return ctrl.Result{}, startErr
	}

	// CreateOrUpdate deployment for publisher proxy
	publisherDeploy, err := r.CreateOrUpdatePublisherProxy(ctx, r.backendType)
	if err != nil {
		backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonPublisherProxySyncFailed, err.Error())
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to update status while creating/updating Event Publisher")
		}
		return ctrl.Result{}, err
	}

	if r.bebSubMgrStarted && !backendStatus.IsSubscriptionControllerStatusReady() {
		backendStatus.SetSubscriptionControllerReadyCondition(true, eventingv1alpha1.ConditionReasonSubscriptionControllerReady, "")
	}

	// CreateOrUpdate status of the CR
	err = r.syncBackendStatus(ctx, backendStatus, publisherDeploy)
	if err != nil {
		return ctrl.Result{}, xerrors.Errorf("failed to create/update %s EventingBackend status: %v", r.backendType, err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) syncOauth2ClientIDAndSecret(ctx context.Context, backendStatus *eventingv1alpha1.EventingBackendStatus) error {
	// Following could return an error when the OAuth2Client CR is created for the first time, until the secret is
	// created by the Hydra operator. However, eventually it should get resolved in the next few reconciliation loops.
	credentials, err := r.getOAuth2ClientCredentials(ctx)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	oauth2CredentialsNotFound := k8serrors.IsNotFound(err)
	oauth2CredentialsChanged := false
	if err == nil && r.isOauth2CredentialsInitialized() {
		oauth2CredentialsChanged = !bytes.Equal(r.credentials.clientID, credentials.clientID) ||
			!bytes.Equal(r.credentials.clientSecret, credentials.clientSecret) ||
			!bytes.Equal(r.credentials.tokenURL, credentials.tokenURL)
	}
	if oauth2CredentialsNotFound || oauth2CredentialsChanged {
		// Stop the controller and mark all subs as not ready
		message := "Stopping the BEB subscription manager due to change in OAuth2 credentials"
		r.namedLogger().Info(message)
		if err := r.bebSubMgr.Stop(false); err != nil {
			return err
		}
		r.bebSubMgrStarted = false
		// update eventing backend status to reflect that the controller is not ready
		backendStatus.SetSubscriptionControllerReadyCondition(false, eventingv1alpha1.ConditionReasonSubscriptionControllerNotReady, message)
		if updateErr := r.syncBackendStatus(ctx, backendStatus, nil); updateErr != nil {
			return errors.Wrapf(err, "update status after stopping BEB controller failed")
		}
	}
	if oauth2CredentialsNotFound {
		return err
	}
	if oauth2CredentialsChanged || !r.isOauth2CredentialsInitialized() {
		r.credentials.clientID = credentials.clientID
		r.credentials.clientSecret = credentials.clientSecret
		r.credentials.tokenURL = credentials.tokenURL
		r.credentials.certsURL = credentials.certsURL
	}
	return nil
}

func setUpEnvironmentForBEBController(secret *v1.Secret) error {
	err := os.Setenv("BEB_API_URL", fmt.Sprintf("%s%s", string(secret.Data[PublisherSecretEMSHostKey]), BEBPublishEndpointForSubscriber))
	if err != nil {
		return errors.Wrapf(err, "set BEB_API_URL env var failed")
	}

	err = os.Setenv("CLIENT_ID", string(secret.Data[deployment.PublisherSecretClientIDKey]))
	if err != nil {
		return errors.Wrapf(err, "set CLIENT_ID env var failed")
	}

	err = os.Setenv("CLIENT_SECRET", string(secret.Data[deployment.PublisherSecretClientSecretKey]))
	if err != nil {
		return errors.Wrapf(err, "set CLIENT_SECRET env var failed")
	}

	err = os.Setenv("TOKEN_ENDPOINT", string(secret.Data[deployment.PublisherSecretTokenEndpointKey]))
	if err != nil {
		return errors.Wrapf(err, "set TOKEN_ENDPOINT env var failed")
	}

	err = os.Setenv("BEB_NAMESPACE", fmt.Sprintf("%s%s", NamespacePrefix, string(secret.Data[deployment.PublisherSecretBEBNamespaceKey])))
	if err != nil {
		return errors.Wrapf(err, "set BEB_NAMESPACE env var failed")
	}

	return nil
}

func (r *Reconciler) syncBackendStatus(ctx context.Context, backendStatus *eventingv1alpha1.EventingBackendStatus, publisher *appsv1.Deployment) error {
	currentBackend, err := r.getCurrentBackendCR(ctx)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "failed to get current EventingBackend")
	}

	// Once backend changes, the publisher deployment changes are not picked up immediately
	if hasBackendTypeChanged(currentBackend.Status, *backendStatus) {
		backendStatus.SetSubscriptionControllerReadyCondition(false,
			eventingv1alpha1.ConditionReasonSubscriptionControllerNotReady, "")
	}

	var publisherReady bool
	if publisher == nil {
		if backendStatus.IsPublisherStatusReady() {
			backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonPublisherDeploymentNotReady, "")
		}
	} else {
		publisherReady = r.isPublisherDeploymentReady(publisher)
		if !publisherReady {
			backendStatus.SetPublisherReadyCondition(false, eventingv1alpha1.ConditionReasonPublisherDeploymentNotReady, "")
		} else {
			backendStatus.SetPublisherReadyCondition(publisherReady, eventingv1alpha1.ConditionReasonPublisherDeploymentReady, "")
		}
	}
	// mark eventing as ready if subscription controller and publisher are ready
	backendStatus.EventingReady = utils.BoolPtr(backendStatus.IsSubscriptionControllerStatusReady() && publisherReady)
	return r.updateStatusAndEmitEvent(ctx, currentBackend, backendStatus)
}

func (r *Reconciler) updateStatusAndEmitEvent(ctx context.Context, currentBackend *eventingv1alpha1.EventingBackend, newBackendStatus *eventingv1alpha1.EventingBackendStatus) error {
	if object.IsBackendStatusEqual(currentBackend.Status, *newBackendStatus) {
		return nil
	}

	// Applying existing attributes
	desiredBackend := currentBackend.DeepCopy()
	desiredBackend.Status = *newBackendStatus

	if err := r.Client.Status().Update(ctx, desiredBackend); err != nil {
		return xerrors.Errorf("failed to update %s EventingBackend status: %v", r.backendType, err)
	}

	// emit event
	r.emitConditionEvents(currentBackend, desiredBackend)

	return nil
}

// emitConditionEvents check each condition, if the condition is modified then emit an event.
func (r *Reconciler) emitConditionEvents(currentBackend, newBackend *eventingv1alpha1.EventingBackend) {
	for _, newCondition := range newBackend.Status.Conditions {
		currentCondition := currentBackend.Status.FindCondition(newCondition.Type)
		if currentCondition != nil && eventingv1alpha1.ConditionEquals(*currentCondition, newCondition) {
			continue
		}
		// condition is modified, so emit an event
		r.emitConditionEvent(newBackend, newCondition)
	}
}

// emitConditionEvent emits a kubernetes event and sets the event type based on the Condition status.
func (r *Reconciler) emitConditionEvent(backend *eventingv1alpha1.EventingBackend, condition eventingv1alpha1.Condition) {
	eventType := v1.EventTypeNormal
	if condition.Status == v1.ConditionFalse {
		eventType = v1.EventTypeWarning
	}
	r.record.Event(backend, eventType, string(condition.Reason), condition.Message)
}

// check if the publisher deployment's pods are ready.
func (r *Reconciler) isPublisherDeploymentReady(publisher *appsv1.Deployment) bool {
	result := *publisher.Spec.Replicas == publisher.Status.ReadyReplicas
	if !result {
		r.namedLogger().Debugf("Event Publisher deployment not ready: expected replicas: %d, got: %d", *publisher.Spec.Replicas, publisher.Status.ReadyReplicas)
	}
	return result
}

func hasBackendTypeChanged(currentBackendStatus, desiredBackendStatus eventingv1alpha1.EventingBackendStatus) bool {
	return currentBackendStatus.Backend != desiredBackendStatus.Backend
}

// getDefaultBackendStatus sets all the conditions and the eventingReady status to true.
func getDefaultBackendStatus() eventingv1alpha1.EventingBackendStatus {
	defaultStatus := eventingv1alpha1.EventingBackendStatus{}
	defaultStatus.InitializeConditions()
	defaultStatus.BEBSecretName = ""
	defaultStatus.BEBSecretNamespace = ""
	defaultStatus.EventingReady = utils.BoolPtr(true)
	return defaultStatus
}

func (r *Reconciler) deletePublisherProxySecret(ctx context.Context) error {
	secretNamespacedName := types.NamespacedName{
		Namespace: deployment.PublisherNamespace,
		Name:      deployment.PublisherName,
	}
	return r.deleteSecret(ctx, secretNamespacedName)
}

func (r *Reconciler) deleteNATSSecret(ctx context.Context) error {
	secretNamespacedName := types.NamespacedName{
		Namespace: kymaSystemNamespace,
		Name:      natsSecretName,
	}
	return r.deleteSecret(ctx, secretNamespacedName)
}

func (r *Reconciler) deleteSecret(ctx context.Context, secretNamespacedName types.NamespacedName) error {
	currentSecret := new(v1.Secret)
	err := r.Get(ctx, secretNamespacedName, currentSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Nothing needs to be done
			return nil
		}
		return err
	}

	if err := r.Client.Delete(ctx, currentSecret); err != nil {
		return errors.Wrapf(err, "failed to delete secret: %s", secretNamespacedName.Name)
	}
	return nil
}

func (r *Reconciler) SyncPublisherProxySecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	secretNamespacedName := types.NamespacedName{
		Namespace: deployment.PublisherNamespace,
		Name:      deployment.PublisherName,
	}
	currentSecret := new(v1.Secret)

	desiredSecret, err := getSecretForPublisher(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid secret for Event Publisher")
	}
	err = r.Get(ctx, secretNamespacedName, currentSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create secret
			r.namedLogger().Debug("Creating secret for BEB publisher")
			err := r.Create(ctx, desiredSecret)
			if err != nil {
				return nil, errors.Wrapf(err, "create secret for Event Publisher failed")
			}
			return desiredSecret, nil
		}
		return nil, errors.Wrapf(err, "Failed to get Event Publisher secret failed")
	}

	if object.Semantic.DeepEqual(currentSecret, desiredSecret) {
		r.namedLogger().Debug("No need to update secret for BEB Event Publisher")
		return currentSecret, nil
	}

	// Update secret
	desiredSecret.ResourceVersion = currentSecret.ResourceVersion
	if err := r.Update(ctx, desiredSecret); err != nil {
		return nil, xerrors.Errorf("failed to update Event Publisher secret: %v", err)
	}

	return desiredSecret, nil
}

func newSecret(name, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func getSecretForPublisher(bebSecret *v1.Secret) (*v1.Secret, error) {
	secret := newSecret(deployment.PublisherName, deployment.PublisherNamespace)

	secret.Labels = map[string]string{
		deployment.AppLabelKey: AppLabelValue,
	}

	if _, ok := bebSecret.Data["messaging"]; !ok {
		return nil, errors.New("message is missing from BEB secret")
	}
	messagingBytes := bebSecret.Data["messaging"]

	if _, ok := bebSecret.Data["namespace"]; !ok {
		return nil, errors.New("namespace is missing from BEB secret")
	}
	namespaceBytes := bebSecret.Data["namespace"]

	var messages []Message
	err := json.Unmarshal(messagingBytes, &messages)
	if err != nil {
		return nil, err
	}

	for _, m := range messages {
		if m.Broker.BrokerType == "saprestmgw" {
			if len(m.OA2.ClientID) == 0 {
				return nil, errors.New("client ID is missing")
			}
			if len(m.OA2.ClientSecret) == 0 {
				return nil, errors.New("client secret is missing")
			}
			if len(m.OA2.TokenEndpoint) == 0 {
				return nil, errors.New("tokenendpoint is missing")
			}
			if len(m.OA2.GrantType) == 0 {
				return nil, errors.New("granttype is missing")
			}
			if len(m.URI) == 0 {
				return nil, errors.New("publish URL is missing")
			}

			secret.StringData = getSecretStringData(m.OA2.ClientID, m.OA2.ClientSecret, m.OA2.TokenEndpoint, m.OA2.GrantType, m.URI, string(namespaceBytes))
			break
		}
	}

	return secret, nil
}

func getSecretStringData(clientID, clientSecret, tokenEndpoint, grantType, publishURL, namespace string) map[string]string {
	return map[string]string{
		deployment.PublisherSecretClientIDKey:      clientID,
		deployment.PublisherSecretClientSecretKey:  clientSecret,
		deployment.PublisherSecretTokenEndpointKey: fmt.Sprintf(TokenEndpointFormat, tokenEndpoint, grantType),
		deployment.PublisherSecretEMSURLKey:        fmt.Sprintf("%s%s", publishURL, BEBPublishEndpointForPublisher),
		PublisherSecretEMSHostKey:                  publishURL,
		deployment.PublisherSecretBEBNamespaceKey:  namespace,
	}
}

func (r *Reconciler) CreateOrUpdatePublisherProxy(ctx context.Context, backend eventingv1alpha1.BackendType) (*appsv1.Deployment, error) {
	return r.CreateOrUpdatePublisherProxyDeployment(ctx, backend, true)
}

func (r *Reconciler) CreateOrUpdatePublisherProxyDeployment(
	ctx context.Context,
	backend eventingv1alpha1.BackendType,
	setOwnerReference bool) (*appsv1.Deployment, error) {
	var desiredPublisher *appsv1.Deployment
	// set backend type here so that the function can be used in eventing-manager
	r.backendType = backend

	switch backend {
	case eventingv1alpha1.NatsBackendType:
		desiredPublisher = deployment.NewNATSPublisherDeployment(r.natsConfig, r.cfg.PublisherConfig)
	case eventingv1alpha1.BEBBackendType:
		desiredPublisher = deployment.NewBEBPublisherDeployment(r.cfg.PublisherConfig)
	default:
		return nil, fmt.Errorf("unknown EventingBackend type %q", backend)
	}

	if setOwnerReference {
		if err := r.setAsOwnerReference(ctx, desiredPublisher); err != nil {
			return nil, errors.Wrapf(err, "set owner reference for Event Publisher failed")
		}
	}

	currentPublisher, err := r.getEPPDeployment(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get Event Publisher deployment")
	}

	if currentPublisher == nil { // no deployment found
		// delete the publisher proxy with invalid backend type if it still exists
		if err := r.deletePublisherProxy(ctx); err != nil {
			return nil, err
		}
		// Create
		r.namedLogger().Debug("Creating Event Publisher deployment")
		return desiredPublisher, r.Create(ctx, desiredPublisher)
	}

	desiredPublisher.ResourceVersion = currentPublisher.ResourceVersion

	// preserve only allowed annotations
	desiredPublisher.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	for k, v := range currentPublisher.Spec.Template.ObjectMeta.Annotations {
		if _, ok := allowedAnnotations[k]; ok {
			desiredPublisher.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}

	if object.Semantic.DeepEqual(currentPublisher, desiredPublisher) {
		return currentPublisher, nil
	}

	// Update publisher proxy deployment
	if err := r.Update(ctx, desiredPublisher); err != nil {
		return nil, errors.Wrapf(err, "update Event Publisher deployment failed")
	}

	return desiredPublisher, nil
}

func (r *Reconciler) CreateOrUpdateBackendCR(ctx context.Context) error {
	labels := map[string]string{
		BackendCRLabelKey: BackendCRLabelValue,
	}
	desiredBackend := &eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.cfg.BackendCRName,
			Namespace: r.cfg.BackendCRNamespace,
			Labels:    labels,
		},
		Spec: eventingv1alpha1.EventingBackendSpec{},
	}

	currentBackend, err := r.getCurrentBackendCR(ctx)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if err := r.Create(ctx, desiredBackend); err != nil {
				return errors.Wrapf(err, "create EventingBackend failed")
			}
			r.namedLogger().Debug("Created EventingBackend")
			return nil
		}
		return errors.Wrapf(err, "get EventingBackend failed")
	}

	desiredBackend.ResourceVersion = currentBackend.ResourceVersion
	if object.Semantic.DeepEqual(&currentBackend, &desiredBackend) {
		return nil
	}

	if err := r.Update(ctx, desiredBackend); err != nil {
		return errors.Wrapf(err, "update EventingBackend failed")
	}

	return nil
}

func (r *Reconciler) getCurrentBackendCR(ctx context.Context) (*eventingv1alpha1.EventingBackend, error) {
	backend := new(eventingv1alpha1.EventingBackend)
	err := r.Get(ctx, types.NamespacedName{
		Name:      r.cfg.BackendCRName,
		Namespace: r.cfg.BackendCRNamespace,
	}, backend)
	return backend, err
}

func (r *Reconciler) getOAuth2SecretNamespacedName() types.NamespacedName {
	var name, namespace string
	if featureflags.IsEventingWebhookAuthEnabled() {
		name = r.cfg.EventingWebhookAuthSecretName
		namespace = r.cfg.EventingWebhookAuthSecretNamespace
	} else {
		name = getOAuth2ClientSecretName()
		namespace = deployment.ControllerNamespace
	}
	return types.NamespacedName{Name: name, Namespace: namespace}
}

func (r *Reconciler) getOAuth2ClientCredentials(ctx context.Context) (*oauth2Credentials, error) {
	var err error
	var exists bool
	var clientID, clientSecret, tokenURL, certsURL []byte

	oauth2Secret := new(v1.Secret)
	oauth2SecretNamespacedName := r.getOAuth2SecretNamespacedName()

	r.namedLogger().Infof("Reading secret %s", oauth2SecretNamespacedName.String())

	if getErr := r.Get(ctx, oauth2SecretNamespacedName, oauth2Secret); getErr != nil {
		err = errors.Wrapf(getErr, "get secret failed namespace:%s name:%s",
			oauth2SecretNamespacedName.Namespace, oauth2SecretNamespacedName.Name)
		return nil, err
	}

	if clientID, exists = oauth2Secret.Data[secretKeyClientID]; !exists {
		err = errors.Errorf("key '%s' not found in secret %s",
			secretKeyClientID, oauth2SecretNamespacedName.String())
		return nil, err
	}

	if clientSecret, exists = oauth2Secret.Data[secretKeyClientSecret]; !exists {
		err = errors.Errorf("key '%s' not found in secret %s",
			secretKeyClientSecret, oauth2SecretNamespacedName.String())
		return nil, err
	}

	if !featureflags.IsEventingWebhookAuthEnabled() {
		tokenURL = []byte(r.envCfg.WebhookTokenEndpoint)
		certsURL = []byte("")
		credentials := oauth2Credentials{
			clientID:     clientID,
			clientSecret: clientSecret,
			tokenURL:     tokenURL,
			certsURL:     certsURL,
		}
		return &credentials, nil
	}

	if tokenURL, exists = oauth2Secret.Data[secretKeyTokenURL]; !exists {
		err = errors.Errorf("key '%s' not found in secret %s",
			secretKeyTokenURL, oauth2SecretNamespacedName.String())
		return nil, err
	}

	if certsURL, exists = oauth2Secret.Data[secretKeyCertsURL]; !exists {
		err = errors.Errorf("key '%s' not found in secret %s",
			secretKeyCertsURL, oauth2SecretNamespacedName.String())
		return nil, err
	}

	credentials := oauth2Credentials{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		certsURL:     certsURL,
	}

	return &credentials, nil
}

func getDeploymentMapper() handler.EventHandler {
	var mapper handler.MapFunc = func(obj client.Object) []reconcile.Request {
		var reqs []reconcile.Request
		// Ignore deployments other than publisher-proxy
		if obj.GetName() == deployment.PublisherName && obj.GetNamespace() == deployment.PublisherNamespace {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: obj.GetNamespace(), Name: "any"},
			})
		}
		return reqs
	}
	return handler.EnqueueRequestsFromMapFunc(mapper)
}

func getEventingBackendCRMapper() handler.EventHandler {
	var mapper handler.MapFunc = func(obj client.Object) []reconcile.Request {
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{Namespace: obj.GetNamespace(), Name: "any"}},
		}
	}
	return handler.EnqueueRequestsFromMapFunc(mapper)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		Watches(&source.Kind{Type: &eventingv1alpha1.EventingBackend{}}, getEventingBackendCRMapper()).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, getDeploymentMapper()).
		Complete(r)
}

func (r *Reconciler) createNATSSecret(ctx context.Context) error {
	secretNamespacedName := types.NamespacedName{
		Namespace: kymaSystemNamespace,
		Name:      natsSecretName,
	}
	currentSecret := new(v1.Secret)
	if err := r.Get(ctx, secretNamespacedName, currentSecret); err != nil {
		if k8serrors.IsNotFound(err) {
			if createErr := r.Client.Create(ctx, constructNATSSecret()); createErr != nil {
				return errors.Wrapf(createErr, "failed to create NATS Secret")
			}
			return nil
		}
		return err
	}
	return nil
}

func (r *Reconciler) startNATSController() error {
	if !r.natsSubMgrStarted {
		if err := r.natsSubMgr.Start(r.cfg.DefaultSubscriptionConfig, subscriptionmanager.Params{}); err != nil {
			return xerrors.Errorf("failed to start NATS subscription manager: %v", err)
		}
		r.natsSubMgrStarted = true
		r.namedLogger().Info("NATS subscription manager was started")
	}
	return nil
}

func (r *Reconciler) stopNATSController() error {
	if r.natsSubMgrStarted {
		if err := r.natsSubMgr.Stop(true); err != nil {
			return xerrors.Errorf("failed to stop NATS subscription manager: %v", err)
		}
		r.natsSubMgrStarted = false
		r.namedLogger().Info("NATS subscription manager was stopped")
	}
	return nil
}

func (r *Reconciler) startBEBController() error {
	if !r.bebSubMgrStarted {
		bebSubMgrParams := subscriptionmanager.Params{
			subscriptionmanager.ParamNameClientID:     r.credentials.clientID,
			subscriptionmanager.ParamNameClientSecret: r.credentials.clientSecret,
			subscriptionmanager.ParamNameTokenURL:     r.credentials.tokenURL,
			subscriptionmanager.ParamNameCertsURL:     r.credentials.certsURL,
		}
		if err := r.bebSubMgr.Start(r.cfg.DefaultSubscriptionConfig, bebSubMgrParams); err != nil {
			return xerrors.Errorf("failed to start BEB subscription manager: %v", err)
		}
		r.bebSubMgrStarted = true
		r.namedLogger().Info("BEB subscription manager was started")
	}
	return nil
}

func (r *Reconciler) stopBEBController() error {
	if r.bebSubMgrStarted {
		if err := r.bebSubMgr.Stop(true); err != nil {
			return xerrors.Errorf("failed to stop BEB subscription manager: %v", err)
		}
		r.bebSubMgrStarted = false
		r.namedLogger().Info("BEB subscription manager was stopped")
	}
	return nil
}

// getEPPDeployment fetches the event publisher by the current active backend type.
func (r *Reconciler) getEPPDeployment(ctx context.Context) (*appsv1.Deployment, error) {
	var list appsv1.DeploymentList
	if err := r.List(ctx, &list, client.MatchingLabels{
		deployment.AppLabelKey:       deployment.PublisherName,
		deployment.InstanceLabelKey:  deployment.InstanceLabelValue,
		deployment.DashboardLabelKey: deployment.DashboardLabelValue,
		deployment.BackendLabelKey:   fmt.Sprint(r.backendType),
	}); err != nil {
		return nil, err
	}

	if len(list.Items) == 0 { // no deployment found
		return nil, nil
	}
	return &list.Items[0], nil
}

// deletePublisherProxy removes the existing publisher proxy.
func (r *Reconciler) deletePublisherProxy(ctx context.Context) error {
	publisherNamespacedName := types.NamespacedName{
		Namespace: deployment.PublisherNamespace,
		Name:      deployment.PublisherName,
	}
	publisher := new(appsv1.Deployment)
	err := r.Get(ctx, publisherNamespacedName, publisher)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	r.namedLogger().Debug("Event Publisher with invalid backend type found, deleting it")
	err = r.Delete(ctx, publisher)
	return err
}

func (r *Reconciler) updateMutatingValidatingWebhookWithCABundle(ctx context.Context) error {
	// get the secret containing the certificate
	var certificateSecret v1.Secret
	secretKey := client.ObjectKey{
		Namespace: kymaSystemNamespace,
		Name:      r.cfg.WebhookSecretName,
	}
	if err := r.Client.Get(ctx, secretKey, &certificateSecret); err != nil {
		return pkgerrors.MakeError(errObjectNotFound, err)
	}

	// get the mutating and validation WH config
	mutatingWH, validatingWH, err := r.getMutatingAndValidatingWebHookConfig(ctx)
	if err != nil {
		return err
	}

	// check that the mutating and validation WH config are valid
	if len(mutatingWH.Webhooks) == 0 {
		return pkgerrors.MakeError(errInvalidObject,
			errors.Errorf("mutatingWH %s does not have associated webhooks", r.cfg.MutatingWebhookName))
	}
	if len(validatingWH.Webhooks) == 0 {
		return pkgerrors.MakeError(errInvalidObject,
			errors.Errorf("validatingWH %s does not have associated webhooks", r.cfg.ValidatingWebhookName))
	}

	// check if the CABundle present is valid
	if !(mutatingWH.Webhooks[0].ClientConfig.CABundle != nil &&
		bytes.Equal(mutatingWH.Webhooks[0].ClientConfig.CABundle, certificateSecret.Data[tlsCertField])) {
		// update the ClientConfig for mutating WH config
		mutatingWH.Webhooks[0].ClientConfig.CABundle = certificateSecret.Data[tlsCertField]
		err = r.Client.Update(ctx, mutatingWH)
		if err != nil {
			return errors.Wrap(err, "while updating mutatingWH with caBundle")
		}
	}

	if !(validatingWH.Webhooks[0].ClientConfig.CABundle != nil &&
		bytes.Equal(validatingWH.Webhooks[0].ClientConfig.CABundle, certificateSecret.Data[tlsCertField])) {
		// update the ClientConfig for validating WH config
		validatingWH.Webhooks[0].ClientConfig.CABundle = certificateSecret.Data[tlsCertField]
		err = r.Client.Update(ctx, validatingWH)
		if err != nil {
			return errors.Wrap(err, "while updating validatingWH with caBundle")
		}
	}

	return nil
}

func (r *Reconciler) getMutatingAndValidatingWebHookConfig(ctx context.Context) (
	*admissionv1.MutatingWebhookConfiguration, *admissionv1.ValidatingWebhookConfiguration, error) {
	var mutatingWH admissionv1.MutatingWebhookConfiguration
	mutatingWHKey := client.ObjectKey{
		Name: r.cfg.MutatingWebhookName,
	}
	if err := r.Client.Get(ctx, mutatingWHKey, &mutatingWH); err != nil {
		return nil, nil, pkgerrors.MakeError(errObjectNotFound, err)
	}
	var validatingWH admissionv1.ValidatingWebhookConfiguration
	validatingWHKey := client.ObjectKey{
		Name: r.cfg.ValidatingWebhookName,
	}
	if err := r.Client.Get(ctx, validatingWHKey, &validatingWH); err != nil {
		return nil, nil, pkgerrors.MakeError(errObjectNotFound, err)
	}
	return &mutatingWH, &validatingWH, nil
}

func (r *Reconciler) isOauth2CredentialsInitialized() bool {
	if featureflags.IsEventingWebhookAuthEnabled() {
		return len(r.credentials.clientID) > 0 &&
			len(r.credentials.clientSecret) > 0 &&
			len(r.credentials.tokenURL) > 0 &&
			len(r.credentials.certsURL) > 0
	}
	return len(r.credentials.clientID) > 0 && len(r.credentials.clientSecret) > 0
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName).With("backend", r.backendType)
}

func getOAuth2ClientSecretName() string {
	return deployment.ControllerName + BEBSecretNameSuffix
}

func constructNATSSecret() *v1.Secret {
	secretMap := make(map[string][]byte)
	password := utils.GetRandString(natsSecretPasswordLength)
	secretMap[natsSecretKey] = []byte(fmt.Sprintf(
		`accounts: {
  "$SYS": {
    users: [
      {user: "admin", password: "%v" }
    ]
  },
}
system_account: "$SYS"`, password))
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      natsSecretName,
			Namespace: kymaSystemNamespace,
			Annotations: map[string]string{
				"eventing.kyma-project.io/managed-by-reconciler-disclaimer": "DO NOT EDIT - " +
					"This resource is managed by Kyma.\n Any modifications breaks eventing.",
			},
		},
		Data: secretMap,
	}
}
