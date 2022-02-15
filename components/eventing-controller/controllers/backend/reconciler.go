package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

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
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
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
)

var (
	// allowedAnnotations are the publisher proxy deployment spec template annotations
	// which should be preserved during reconciliation.
	allowedAnnotations = map[string]string{
		"kubectl.kubernetes.io/restartedAt": "",
	}
)

type Reconciler struct {
	client.Client
	ctx               context.Context
	natsSubMgr        subscriptionmanager.Manager
	natsSubMgrStarted bool
	bebSubMgr         subscriptionmanager.Manager
	bebSubMgrStarted  bool
	logger            *logger.Logger
	record            record.EventRecorder
	cfg               env.BackendConfig
	// backendType is the type of the backend which the reconciler detects at runtime
	backendType eventingv1alpha1.BackendType
	// The OAuth2 credentials that are passed to the BEB subscription reconciler
	oauth2ClientID     []byte
	oauth2ClientSecret []byte
}

func NewReconciler(ctx context.Context, natsSubMgr, bebSubMgr subscriptionmanager.Manager, client client.Client, logger *logger.Logger, recorder record.EventRecorder) *Reconciler {
	cfg := env.GetBackendConfig()
	return &Reconciler{
		ctx:        ctx,
		natsSubMgr: natsSubMgr,
		bebSubMgr:  bebSubMgr,
		Client:     client,
		logger:     logger,
		record:     recorder,
		cfg:        cfg,
	}
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete

func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	var secretList v1.SecretList

	if err := r.List(ctx, &secretList, client.MatchingLabels{
		BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
	}); err != nil {
		return ctrl.Result{}, err
	}

	if len(secretList.Items) > 1 {
		// This is not allowed!
		r.namedLogger().Debugw("more than one secret with the label exist", "key", BEBBackendSecretLabelKey, "value", BEBBackendSecretLabelValue, "count", len(secretList.Items))
		currentBackend, err := r.getCurrentBackendCR(ctx)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		desiredBackend := currentBackend.DeepCopy()
		desiredBackend.Status = getDefaultBackendStatus()
		// Do not change the value of backend type if it cannot be changed
		desiredBackend.Status.Backend = currentBackend.Status.Backend
		desiredBackend.Status.EventingReady = utils.BoolPtr(false)
		if object.Semantic.DeepEqual(&desiredBackend.Status, &currentBackend.Status) {
			return ctrl.Result{}, nil
		}
		if err := r.Client.Status().Update(ctx, desiredBackend); err != nil {
			r.namedLogger().Errorw("update EventingBackend status failed", "error", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// If secret with label then BEB flow
	if len(secretList.Items) == 1 {
		return r.reconcileBEBBackend(ctx, &secretList.Items[0])
	}

	// Default: NATS flow
	return r.reconcileNATSBackend(ctx)
}

func (r *Reconciler) reconcileNATSBackend(ctx context.Context) (ctrl.Result, error) {
	// CreateOrUpdate CR with NATS
	r.backendType = eventingv1alpha1.NatsBackendType
	newBackend, err := r.CreateOrUpdateBackendCR(ctx)
	if err != nil {
		// Update status if bad
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, nil, nil, nil)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when create or update backend failed")
		}
		return ctrl.Result{}, errors.Wrapf(err, "create or update EventingBackend failed, type: %s", eventingv1alpha1.NatsBackendType)
	}

	// Stop the BEB subscription controller
	if err := r.stopBEBController(); err != nil {
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, nil)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when stop BEB controller failed")
		}
		return ctrl.Result{}, err
	}
	// Start the NATS subscription controller
	if err := r.startNATSController(); err != nil {
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, nil)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when start NATS controller failed")
		}
		return ctrl.Result{}, err
	}

	// Delete secret for publisher proxy if it exists
	err = r.DeletePublisherProxySecret(ctx)
	if err != nil {
		// Update status if bad
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, nil)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when delete publisher proxy secret failed")
		}
		return ctrl.Result{}, errors.Wrapf(err, "delete eventing publisher proxy secret failed")
	}

	// CreateOrUpdate deployment for publisher proxy
	publisher, err := r.CreateOrUpdatePublisherProxy(ctx, r.backendType)
	if err != nil {
		// Update status if bad
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, nil)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when create or update publisher proxy failed")
		}
		r.namedLogger().Errorw("create or update eventing publisher proxy deployment failed", "error", err)
		return ctrl.Result{}, err
	}

	// CreateOrUpdate status of the CR
	// Get publisher proxy ready status
	err = r.UpdateBackendStatus(ctx, r.backendType, newBackend, publisher, nil)
	return ctrl.Result{}, err
}

func (r *Reconciler) reconcileBEBBackend(ctx context.Context, bebSecret *v1.Secret) (ctrl.Result, error) {
	r.backendType = eventingv1alpha1.BEBBackendType
	// CreateOrUpdate CR with BEB
	newBackend, err := r.CreateOrUpdateBackendCR(ctx)
	if err != nil {
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, nil, nil, bebSecret)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when create or update backend CR failed")
		}
		return ctrl.Result{}, errors.Wrapf(err, "create or update EventingBackend failed, type: %s", eventingv1alpha1.BEBBackendType)
	}

	// Stop the NATS subscription controller
	if err := r.stopNATSController(); err != nil {
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, bebSecret)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when stop NATS controller failed")
		}
		return ctrl.Result{}, err
	}

	// Following could return an error when the OAuth2Client CR is created for the first time, until the secret is
	// created by the Hydra operator. However, eventually it should get resolved in the next few reconciliation loops.
	oauth2ClientID, oauth2ClientSecret, err := r.getOAuth2ClientCredentials(ctx, getOAuth2ClientSecretName(), deployment.ControllerNamespace)
	if err != nil && !k8serrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	oauth2CredentialsNotFound := k8serrors.IsNotFound(err)
	oauth2CredentialsChanged := false
	if err == nil {
		oauth2CredentialsChanged = !bytes.Equal(r.oauth2ClientID, oauth2ClientID) || !bytes.Equal(r.oauth2ClientSecret, oauth2ClientSecret)
	}
	if oauth2CredentialsNotFound || oauth2CredentialsChanged {
		// Stop the controller and mark all subs as not ready
		r.namedLogger().Info("stopping the BEB subscription manager due to change in OAuth2 credentials")
		if err := r.bebSubMgr.Stop(false); err != nil {
			return ctrl.Result{}, err
		}
		r.bebSubMgrStarted = false
		if updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, bebSecret); updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status after stopping BEB controller failed")
		}
	}
	if oauth2CredentialsNotFound {
		return ctrl.Result{}, err
	}
	if oauth2CredentialsChanged {
		r.oauth2ClientID = oauth2ClientID
		r.oauth2ClientSecret = oauth2ClientSecret
	}

	// CreateOrUpdate deployment for publisher proxy secret
	secretForPublisher, err := r.SyncPublisherProxySecret(ctx, bebSecret)
	if err != nil {
		// Update status if bad
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, bebSecret)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when sync publisher proxy secret failed")
		}
		r.namedLogger().Errorw("sync publisher proxy secret failed", "backend", eventingv1alpha1.BEBBackendType, "error", err)
		return ctrl.Result{}, err
	}

	// Set environment with secrets for BEB subscription controller
	err = setUpEnvironmentForBEBController(secretForPublisher)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "setup env var for BEB controller failed")
	}

	// Start the BEB subscription controller
	if err := r.startBEBController(oauth2ClientID, oauth2ClientSecret); err != nil {
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, bebSecret)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when start BEB controller failed")
		}
		return ctrl.Result{}, err
	}

	// CreateOrUpdate deployment for publisher proxy
	publisherDeploy, err := r.CreateOrUpdatePublisherProxy(ctx, r.backendType)
	if err != nil {
		// Update status if bad
		updateErr := r.UpdateBackendStatus(ctx, r.backendType, newBackend, nil, bebSecret)
		if updateErr != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status when create or update publisher proxy failed")
		}
		r.namedLogger().Errorw("create or update publisher proxy failed", "backend", r.backendType, "error", err)
		return ctrl.Result{}, err
	}
	// CreateOrUpdate status of the CR
	err = r.UpdateBackendStatus(ctx, r.backendType, newBackend, publisherDeploy, bebSecret)
	if err != nil {
		r.namedLogger().Errorw("create or update backend status failed", "backend", r.backendType, "error", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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

func (r *Reconciler) UpdateBackendStatus(ctx context.Context, backendType eventingv1alpha1.BackendType, backend *eventingv1alpha1.EventingBackend, publisher *appsv1.Deployment, bebSecret *v1.Secret) error {
	var publisherReady, subscriptionControllerReady bool

	currentBackend, err := r.getCurrentBackendCR(ctx)
	if err != nil {
		return errors.Wrapf(err, "get current backend failed")
	}
	currentStatus := currentBackend.Status

	desiredStatus := getDefaultBackendStatus()
	desiredStatus.Backend = backendType

	// When backend creation fails, marking the eventingbackend ready to false
	if backend == nil {
		// Applying existing attributes
		desiredBackend := currentBackend.DeepCopy()
		desiredBackend.Status = desiredStatus

		if object.Semantic.DeepEqual(&desiredStatus, &currentStatus) {
			return nil
		}

		r.namedLogger().Debug("update backend CR status when backend is nil")
		if err := r.Client.Status().Update(ctx, desiredBackend); err != nil {
			r.namedLogger().Errorw("update EventingBackend status failed", "error", err)
			return err
		}
		return nil
	}

	// Once backend changes, the publisher deployment changes are not picked up immediately
	// Hence marking the EventingBackend ready to false
	if hasBackendTypeChanged(currentStatus, desiredStatus) {
		// Applying existing attributes
		desiredBackend := currentBackend.DeepCopy()
		desiredBackend.Status = desiredStatus

		if object.Semantic.DeepEqual(&desiredStatus, &currentStatus) {
			return nil
		}

		if err := r.Client.Status().Update(ctx, desiredBackend); err != nil {
			r.namedLogger().Errorw("update EventingBackend status failed", "error", err)
			return err
		}
		return nil
	}

	publisherReady, err = isPublisherDeploymentReady(ctx, backendType, publisher, r)
	if err != nil {
		return err
	}

	switch backendType {
	case eventingv1alpha1.BEBBackendType:
		if bebSecret != nil {
			desiredStatus.BEBSecretName = bebSecret.Name
			desiredStatus.BEBSecretNamespace = bebSecret.Namespace
		}
		subscriptionControllerReady = r.bebSubMgrStarted
	case eventingv1alpha1.NatsBackendType:
		desiredStatus.BEBSecretName = ""
		desiredStatus.BEBSecretNamespace = ""
		subscriptionControllerReady = r.natsSubMgrStarted
	}
	eventingReady := subscriptionControllerReady && publisherReady

	desiredStatus.Backend = backendType
	desiredStatus.SubscriptionControllerReady = utils.BoolPtr(subscriptionControllerReady)
	desiredStatus.EventingReady = utils.BoolPtr(eventingReady)
	desiredStatus.PublisherProxyReady = utils.BoolPtr(publisherReady)

	if object.Semantic.DeepEqual(&desiredStatus, &currentStatus) {
		return nil
	}

	// Applying existing attributes
	desiredBackend := currentBackend.DeepCopy()
	desiredBackend.Status = desiredStatus

	if err := r.Client.Status().Update(ctx, desiredBackend); err != nil {
		r.namedLogger().Errorw("update EventingBackend status failed", "error", err)
		return err
	}
	return nil
}

// check the publisher proxy deployment pods for the right backend type and ready status
func isPublisherDeploymentReady(ctx context.Context, backendType eventingv1alpha1.BackendType, publisher *appsv1.Deployment, r *Reconciler) (bool, error) {
	if publisher == nil {
		return false, nil
	}

	// get the publisherDeployment's pods
	var pods v1.PodList
	if err := r.List(ctx, &pods, client.MatchingLabels{
		deployment.AppLabelKey: deployment.PublisherName,
	}); err != nil {
		return false, err
	}

	var containers []v1.Container
	var statuses []v1.ContainerStatus
	for _, pod := range pods.Items {
		if pod.DeletionTimestamp.IsZero() {
			containers = append(containers, pod.Spec.Containers...)
			statuses = append(statuses, pod.Status.ContainerStatuses...)
		}
	}

	// check the pods for right backend type
	for _, container := range containers {
		if container.Name == deployment.PublisherName {
			for _, envVar := range container.Env {
				if strings.EqualFold(envVar.Name, "BACKEND") && !strings.EqualFold(envVar.Value, fmt.Sprint(backendType)) {
					return false, nil
				}
			}
		}
	}

	// check if the container's status is set to Ready
	readyPodsCount := 0
	for _, status := range statuses {
		// skip the sidecars
		if status.Name != deployment.PublisherName {
			continue
		}
		if !status.Ready {
			return false, nil
		}
		readyPodsCount++
		// the ready pods number of the right backend type should match the spec
	}
	return readyPodsCount == int(*publisher.Spec.Replicas), nil
}

func hasBackendTypeChanged(currentBackendStatus, desiredBackendStatus eventingv1alpha1.EventingBackendStatus) bool {
	return currentBackendStatus.Backend != desiredBackendStatus.Backend
}

func getDefaultBackendStatus() eventingv1alpha1.EventingBackendStatus {
	return eventingv1alpha1.EventingBackendStatus{
		SubscriptionControllerReady: utils.BoolPtr(false),
		PublisherProxyReady:         utils.BoolPtr(false),
		EventingReady:               utils.BoolPtr(false),
		BEBSecretName:               "",
		BEBSecretNamespace:          "",
	}
}

func (r *Reconciler) DeletePublisherProxySecret(ctx context.Context) error {
	secretNamespacedName := types.NamespacedName{
		Namespace: deployment.PublisherNamespace,
		Name:      deployment.PublisherName,
	}
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
		return errors.Wrapf(err, "delete eventing publisher proxy secret failed")
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
		return nil, errors.Wrapf(err, "invalid secret for publisher")
	}
	err = r.Get(ctx, secretNamespacedName, currentSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create secret
			r.namedLogger().Debug("creating secret for BEB publisher")
			err := r.Create(ctx, desiredSecret)
			if err != nil {
				return nil, errors.Wrapf(err, "create secret for eventing publisher proxy failed")
			}
			return desiredSecret, nil
		}
		return nil, errors.Wrapf(err, "get eventing publisher proxy secret failed")
	}

	if object.Semantic.DeepEqual(currentSecret, desiredSecret) {
		r.namedLogger().Debug("no need to update secret for BEB publisher")
		return currentSecret, nil
	}

	// Update secret
	desiredSecret.ResourceVersion = currentSecret.ResourceVersion
	if err := r.Update(ctx, desiredSecret); err != nil {
		r.namedLogger().Errorw("update publisher proxy secret failed", "error", err)
		return nil, err
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
	publisherNamespacedName := types.NamespacedName{
		Namespace: deployment.PublisherNamespace,
		Name:      deployment.PublisherName,
	}
	currentPublisher := new(appsv1.Deployment)
	var desiredPublisher *appsv1.Deployment

	switch backend {
	case eventingv1alpha1.NatsBackendType:
		desiredPublisher = deployment.NewNATSPublisherDeployment(r.cfg.PublisherConfig)
	case eventingv1alpha1.BEBBackendType:
		desiredPublisher = deployment.NewBEBPublisherDeployment(r.cfg.PublisherConfig)
	default:
		return nil, fmt.Errorf("unknown eventing backend type %q", backend)
	}

	if err := r.setAsOwnerReference(ctx, desiredPublisher); err != nil {
		return nil, errors.Wrapf(err, "set owner reference for publisher failed")
	}

	err := r.Get(ctx, publisherNamespacedName, currentPublisher)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create
			r.namedLogger().Debug("creating publisher proxy")
			return desiredPublisher, r.Create(ctx, desiredPublisher)
		}
		return nil, err
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
		return nil, errors.Wrapf(err, "update publisher proxy deployment failed")
	}

	return desiredPublisher, nil
}

func (r *Reconciler) CreateOrUpdateBackendCR(ctx context.Context) (*eventingv1alpha1.EventingBackend, error) {
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
				return nil, errors.Wrapf(err, "create EventingBackend failed")
			}
			r.namedLogger().Debug("created backend CR")
			return desiredBackend, nil
		}
		return nil, errors.Wrapf(err, "get EventingBackend failed")
	}

	desiredBackend.ResourceVersion = currentBackend.ResourceVersion
	if object.Semantic.DeepEqual(&currentBackend, &desiredBackend) {
		return currentBackend, nil
	}

	if err := r.Update(ctx, desiredBackend); err != nil {
		return nil, errors.Wrapf(err, "update EventingBackend failed")
	}

	return desiredBackend, nil
}

func (r *Reconciler) getCurrentBackendCR(ctx context.Context) (*eventingv1alpha1.EventingBackend, error) {
	backend := new(eventingv1alpha1.EventingBackend)
	err := r.Get(ctx, types.NamespacedName{
		Name:      r.cfg.BackendCRName,
		Namespace: r.cfg.BackendCRNamespace,
	}, backend)
	return backend, err
}

func (r *Reconciler) getOAuth2ClientCredentials(ctx context.Context, name, namespace string) (clientID, clientSecret []byte, err error) {
	oauth2Secret := new(v1.Secret)
	oauth2SecretNamespacedName := types.NamespacedName{Namespace: namespace, Name: name}
	if getErr := r.Get(ctx, oauth2SecretNamespacedName, oauth2Secret); getErr != nil {
		err = errors.Wrapf(getErr, "get secret failed namespace:%s name:%s", namespace, name)
		return
	}
	var exists bool
	if clientID, exists = oauth2Secret.Data["client_id"]; !exists {
		err = errors.New("key 'client_id' not found in secret " + oauth2SecretNamespacedName.String())
		return
	}
	if clientSecret, exists = oauth2Secret.Data["client_secret"]; !exists {
		err = errors.New("key 'client_secret' not found in secret " + oauth2SecretNamespacedName.String())
		return
	}
	return
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

func (r *Reconciler) startNATSController() error {
	if !r.natsSubMgrStarted {
		if err := r.natsSubMgr.Start(r.cfg.DefaultSubscriptionConfig, subscriptionmanager.Params{}); err != nil {
			r.namedLogger().Errorw("start NATS subscription manager failed", "error", err)
			return err
		}
		r.natsSubMgrStarted = true
		r.namedLogger().Info("start NATS subscription manager succeeded")
	}
	return nil
}

func (r *Reconciler) stopNATSController() error {
	if r.natsSubMgrStarted {
		if err := r.natsSubMgr.Stop(true); err != nil {
			r.namedLogger().Errorw("stop NATS subscription manager failed", "error", err)
			return err
		}
		r.natsSubMgrStarted = false
		r.namedLogger().Info("stop NATS subscription manager succeeded")
	}
	return nil
}

func (r *Reconciler) startBEBController(clientID, clientSecret []byte) error {
	if !r.bebSubMgrStarted {
		bebSubMgrParams := subscriptionmanager.Params{"client_id": clientID, "client_secret": clientSecret}
		if err := r.bebSubMgr.Start(r.cfg.DefaultSubscriptionConfig, bebSubMgrParams); err != nil {
			r.namedLogger().Errorw("start BEB subscription manager failed", "error", err)
			return err
		}
		r.bebSubMgrStarted = true
		r.namedLogger().Info("start BEB subscription manager succeeded")
	}
	return nil
}

func (r *Reconciler) stopBEBController() error {
	if r.bebSubMgrStarted {
		if err := r.bebSubMgr.Stop(true); err != nil {
			r.namedLogger().Errorw("stop BEB subscription manager failed", "error", err)
			return err
		}
		r.bebSubMgrStarted = false
		r.namedLogger().Info("stop BEB subscription manager succeeded")
	}
	return nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName).With("backend", r.backendType)
}

func getOAuth2ClientSecretName() string {
	return deployment.ControllerName + BEBSecretNameSuffix
}
