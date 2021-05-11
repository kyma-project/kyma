package backend

import (
	"context"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/go-logr/logr"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BEBBackendSecretLabelKey   = "kyma-project.io/eventing-backend"
	BEBBackendSecretLabelValue = "beb"
	DefaultEventingBackendName = "eventing-backend"
	// TODO: where to get this namespace
	DefaultEventingBackendNamespace = "kyma-system"

	PublisherNamespace  = "kyma-system"
	PublisherName       = "eventing-publisher-proxy"
	BackendCRLabelKey   = "kyma-project.io/eventing"
	BackendCRLabelValue = "backend"
	AppLabelKey         = "app.kubernetes.io/name"
	AppLabelValue       = PublisherName

	TokenEndpointFormat = "%s?grant_type=%s&response_type=token"
)

type Reconciler struct {
	ctx context.Context
	client.Client
	cache.Cache
	Log    logr.Logger
	record record.EventRecorder
}

func NewReconciler(ctx context.Context, client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		ctx:    ctx,
		Client: client,
		Cache:  cache,
		Log:    log,
		record: recorder,
	}
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var secretList v1.SecretList

	if err := r.Cache.List(ctx, &secretList, client.MatchingLabels{
		BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
	}); err != nil {
		return ctrl.Result{}, err
	}
	r.Log.Info("Found secrets with BEB label", "count", len(secretList.Items))

	if len(secretList.Items) > 1 {
		// This is not allowed!
		r.Log.Info(fmt.Sprintf("more than one secret with the label %q=%q exist", BEBBackendSecretLabelKey, BEBBackendSecretLabelValue))
		backend, err := r.getCurrentBackendCR(ctx)
		if err == nil && *backend.Status.EventingReady {
			backend.Status.EventingReady = boolPtr(false)
			err := r.Status().Update(ctx, backend)
			return ctrl.Result{}, err
		}
		if !k8serrors.IsNotFound(err) {
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
	r.Log.Info("Reconciling with backend as NATS")
	backendType := eventingv1alpha1.NatsBackendType

	newBackend, err := r.CreateOrUpdateBackendCR(ctx)
	if err != nil {
		// Update status if bad
		return ctrl.Result{}, errors.Wrapf(err, "failed to createOrUpdate EventingBackend, type: %s", eventingv1alpha1.NatsBackendType)
	}
	r.Log.Info("Created/updated backend CR")

	// TODO: Subscription controller logic
	// Stop subscription controller (Radu/Frank)
	// Start the other subscription controller (Radu/Frank)

	// Delete secret for publisher proxy if it exists
	err = r.DeletePublisherProxySecret(ctx)
	if err != nil {
		// Update status if bad
		return ctrl.Result{}, errors.Wrapf(err, "cannot delete eventing publisher proxy secret")
	}

	// CreateOrUpdate deployment for publisher proxy
	r.Log.Info("trying to create/update eventing publisher proxy...")
	publisher, err := r.CreateOrUpdatePublisherProxy(ctx, backendType)
	if err != nil {
		// Update status if bad
		r.Log.Error(err, "cannot create/update eventing publisher proxy deployment")
		return ctrl.Result{}, err
	}
	r.Log.Info("Created/updated publisher proxy")

	// CreateOrUpdate status of the CR
	// Get publisher proxy ready status
	err = r.UpdateBackendStatus(ctx, backendType, newBackend, publisher, nil)
	return ctrl.Result{}, err
}

func (r *Reconciler) reconcileBEBBackend(ctx context.Context, bebSecret *v1.Secret) (ctrl.Result, error) {
	r.Log.Info("Reconciling with backend as BEB")
	backendType := eventingv1alpha1.BebBackendType
	// CreateOrUpdate CR with BEB
	newBackend, err := r.CreateOrUpdateBackendCR(ctx)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to createOrUpdate EventingBackend, type: %s", eventingv1alpha1.BebBackendType)
	}

	// TODO: Controller logic
	// Stop subscription controller (Radu/Frank)
	// Start the other subscription controller (Radu/Frank)

	// CreateOrUpdate deployment for publisher proxy secret
	_, err = r.SyncPublisherProxySecret(ctx, bebSecret)
	if err != nil {
		// Update status if bad
		r.Log.Error(err, "failed to sync publisher proxy secret", "backend", eventingv1alpha1.BebBackendType)
		return ctrl.Result{}, err
	}

	// CreateOrUpdate deployment for publisher proxy
	publisherDeploy, err := r.CreateOrUpdatePublisherProxy(ctx, backendType)
	if err != nil {
		// Update status if bad
		r.Log.Error(err, "failed to create or update publisher proxy", "backend", backendType)
		return ctrl.Result{}, err
	}
	// CreateOrUpdate status of the CR
	err = r.UpdateBackendStatus(ctx, backendType, newBackend, publisherDeploy, bebSecret)
	if err != nil {
		r.Log.Error(err, "failed to create or update backend status", "backend", backendType)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) UpdateBackendStatus(ctx context.Context, backendType eventingv1alpha1.BackendType, backend *eventingv1alpha1.EventingBackend, publisher *appsv1.Deployment, bebSecret *v1.Secret) error {
	currentBackend, err := r.getCurrentBackendCR(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to get current backend")
	}

	currentStatus := currentBackend.Status

	// In case a publisher already exists, make sure during the switch the status of publisherReady is false
	publisherReady := publisher.Status.Replicas == publisher.Status.ReadyReplicas

	// TODO Subscription controller logic
	subscriptionControllerReady := true

	eventingReady := subscriptionControllerReady && publisherReady

	desiredStatus := eventingv1alpha1.EventingBackendStatus{
		Backend:                     backendType,
		SubscriptionControllerReady: boolPtr(subscriptionControllerReady),
		EventingReady:               boolPtr(eventingReady),
		PublisherProxyReady:         boolPtr(publisherReady),
	}

	switch backendType {
	case eventingv1alpha1.BebBackendType:
		desiredStatus.BebSecretName = bebSecret.Name
		desiredStatus.BebSecretNamespace = bebSecret.Namespace
	case eventingv1alpha1.NatsBackendType:
		desiredStatus.BebSecretName = ""
		desiredStatus.BebSecretNamespace = ""
	}

	if object.Semantic.DeepEqual(&desiredStatus, &currentStatus) {
		r.Log.Info("No need to update backend CR status")
		return nil
	}

	// Applying existing attributes
	desiredBackend := currentBackend.DeepCopy()
	desiredBackend.Status = desiredStatus

	r.Log.Info("Updating backend CR status")
	if err := r.Client.Status().Update(ctx, desiredBackend); err != nil {
		r.Log.Error(err, "error updating EventingBackend status")
		return err
	}
	return nil
}

func (r *Reconciler) DeletePublisherProxySecret(ctx context.Context) error {
	secretNamespacedName := types.NamespacedName{
		Namespace: PublisherNamespace,
		Name:      PublisherName,
	}
	currentSecret := new(v1.Secret)
	err := r.Cache.Get(ctx, secretNamespacedName, currentSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Nothing needs to be done
			return nil
		}
		return err
	}

	if err := r.Client.Delete(ctx, currentSecret); err != nil {
		return errors.Wrapf(err, "failed to delete eventing publisher proxy secret")
	}
	return nil
}

func (r *Reconciler) SyncPublisherProxySecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	secretNamespacedName := types.NamespacedName{
		Namespace: PublisherNamespace,
		Name:      PublisherName,
	}
	currentSecret := new(v1.Secret)

	desiredSecret, err := getSecretForPublisher(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid secret for publisher")
	}
	err = r.Cache.Get(ctx, secretNamespacedName, currentSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create secret
			r.Log.Info("creating secret for BEB publisher")
			err := r.Create(ctx, desiredSecret)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create secret for eventing publisher proxy")
			}
			return desiredSecret, nil
		}
		return nil, errors.Wrapf(err, "failed to get eventing publisher proxy secret")
	}

	if object.Semantic.DeepEqual(currentSecret, desiredSecret) {
		r.Log.Info("No need to update secret for BEB publisher")
		return currentSecret, nil
	}

	// Update secret
	desiredSecret.ResourceVersion = currentSecret.ResourceVersion
	if err := r.Update(ctx, desiredSecret); err != nil {
		r.Log.Error(err, "Cannot update publisher proxy secret")
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
	secret := newSecret(PublisherName, PublisherNamespace)

	secret.Labels = map[string]string{
		AppLabelKey: AppLabelValue,
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
		"client-id":       clientID,
		"client-secret":   clientSecret,
		"token-endpoint":  fmt.Sprintf(TokenEndpointFormat, tokenEndpoint, grantType),
		"ems-publish-url": publishURL,
		"beb-namespace":   namespace,
	}
}

func (r *Reconciler) CreateOrUpdatePublisherProxy(ctx context.Context, backend eventingv1alpha1.BackendType) (*appsv1.Deployment, error) {
	publisherNamespacedName := types.NamespacedName{
		Namespace: PublisherNamespace,
		Name:      PublisherName,
	}
	currentPublisher := new(appsv1.Deployment)
	var desiredPublisher *appsv1.Deployment

	switch backend {
	case eventingv1alpha1.NatsBackendType:
		desiredPublisher = newNATSPublisherDeployment()
	case eventingv1alpha1.BebBackendType:
		desiredPublisher = newBEBPublisherDeployment()
	default:
		return nil, fmt.Errorf("unknown eventing backend type %q", backend)
	}

	err := r.Cache.Get(ctx, publisherNamespacedName, currentPublisher)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create
			r.Log.Info("creating publisher proxy")
			return desiredPublisher, r.Create(ctx, desiredPublisher)
		}
		return nil, err
	}

	desiredPublisher.ResourceVersion = currentPublisher.ResourceVersion
	if object.Semantic.DeepEqual(currentPublisher, desiredPublisher) {
		r.Log.Info("No need to update publisher proxy")
		return currentPublisher, nil
	}

	// Update publisher proxy deployment
	r.Log.Info("updating publisher proxy")
	if err := r.Update(ctx, desiredPublisher); err != nil {
		return nil, errors.Wrapf(err, "cannot update publisher proxy deployment")
	}

	return desiredPublisher, nil
}

func (r *Reconciler) CreateOrUpdateBackendCR(ctx context.Context) (*eventingv1alpha1.EventingBackend, error) {
	labels := map[string]string{
		BackendCRLabelKey: BackendCRLabelValue,
	}
	desiredBackend := &eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultEventingBackendName,
			Namespace: DefaultEventingBackendNamespace,
			Labels:    labels,
		},
		Spec: eventingv1alpha1.EventingBackendSpec{},
		//// TODO: setting status is not needed anymore? Doesn't even work!
		//Status: eventingv1alpha1.EventingBackendStatus{
		//	Backend:                     backend,
		//	EventingReady:               boolPtr(false),
		//	SubscriptionControllerReady: boolPtr(false),
		//	PublisherProxyReady:         boolPtr(false),
		//	BebSecretName:               "",
		//	BebSecretNamespace:          "",
		//},
	}

	currentBackend, err := r.getCurrentBackendCR(ctx)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.Log.Info("trying to create backend CR...")
			if err := r.Create(ctx, desiredBackend); err != nil {
				return nil, errors.Wrapf(err, "cannot create an EventingBackend")
			}
			r.Log.Info("created backend CR")
			return desiredBackend, nil
		}
		return nil, errors.Wrapf(err, "failed to get an EventingBackend")
	}

	r.Log.Info("Found existing backend CR")
	desiredBackend.ResourceVersion = currentBackend.ResourceVersion
	if object.Semantic.DeepEqual(&currentBackend, &desiredBackend) {
		r.Log.Info("No need to update existing backend CR")
		return currentBackend, nil
	}
	r.Log.Info("Update existing backend CR")
	if err := r.Update(ctx, desiredBackend); err != nil {
		return nil, errors.Wrapf(err, "cannot update the EventingBackend")
	}

	return desiredBackend, nil
}

func (r *Reconciler) getCurrentBackendCR(ctx context.Context) (*eventingv1alpha1.EventingBackend, error) {
	backend := new(eventingv1alpha1.EventingBackend)
	err := r.Cache.Get(ctx, types.NamespacedName{
		Namespace: DefaultEventingBackendNamespace,
		Name:      DefaultEventingBackendName,
	}, backend)
	return backend, err
}

func getDeploymentMapper() handler.EventHandler {
	var mapper handler.ToRequestsFunc = func(mo handler.MapObject) []reconcile.Request {
		var reqs []reconcile.Request
		// Ignore deployments other than publisher-proxy
		if mo.Meta.GetName() == PublisherName && mo.Meta.GetNamespace() == PublisherNamespace {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: mo.Meta.GetNamespace(), Name: "any"},
			})
		}
		return reqs
	}
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &mapper}
}

func getEventingBackendCRMapper() handler.EventHandler {
	var mapper handler.ToRequestsFunc = func(mo handler.MapObject) []reconcile.Request {
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{Namespace: mo.Meta.GetNamespace(), Name: "any"}},
		}
	}
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &mapper}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		Watches(&source.Kind{Type: &eventingv1alpha1.EventingBackend{}}, getEventingBackendCRMapper()).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, getDeploymentMapper()).
		Complete(r)
}
