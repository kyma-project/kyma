package httpsource

import (
	"context"
	"reflect"
	"strconv"

	pkgerrors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	messaginglistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	apisv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	securityclientv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"
	securitylistersv1beta1 "istio.io/client-go/pkg/listers/security/v1beta1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	sourceslistersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/errors"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/object"
)

// Reconciler reconciles HTTPSource resources.
type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// adapter properties
	adapterEnvCfg     *httpAdapterEnvConfig
	adapterMetricsCfg string
	adapterLoggingCfg string

	// listers index properties about resources
	httpsourceLister         sourceslistersv1alpha1.HTTPSourceLister
	deploymentLister         appslistersv1.DeploymentLister
	chLister                 messaginglistersv1alpha1.ChannelLister
	serviceLister            v1.ServiceLister
	peerAuthenticationLister securitylistersv1beta1.PeerAuthenticationLister

	// clients allow interactions with API objects
	sourcesClient   sourcesclientv1alpha1.SourcesV1alpha1Interface
	messagingClient messagingclientv1alpha1.MessagingV1alpha1Interface
	securityClient  securityclientv1beta1.SecurityV1beta1Interface

	// URI resolver for sink destinations
	sinkResolver *resolver.URIResolver
}

// Mandatory adapter env vars
const (
	// Common
	// see https://github.com/knative/eventing/blob/release-0.10/pkg/adapter/config.go
	sinkURIEnvVar        = "SINK_URI"
	namespaceEnvVar      = "NAMESPACE"
	metricsConfigEnvVar  = "K_METRICS_CONFIG"
	loggingConfigEnvVar  = "K_LOGGING_CONFIG"
	adapterPortEnvVar    = "PORT"
	adapterContainerName = "source"
	portName             = "http-cloudevent"
	externalPort         = 80
	metricsPort          = 9092
	metricsPortName      = "http-usermetric"

	// HTTP adapter specific
	eventSourceEnvVar = "EVENT_SOURCE"
)

const adapterHealthEndpoint = "/healthz"
const adapterPort = 8080

const (
	applicationNameLabelKey = "application-name"
	applicationLabelKey     = "app"
)

// Reconcile compares the actual state of a HTTPSource object referenced by key
// with its desired state, and attempts to converge the two.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	src, err := httpSourceByKey(key, r.httpsourceLister)
	if err != nil {
		return errors.Handle(err, ctx, "Failed to get object from local store")
	}

	ch, err := r.reconcileSink(src)
	if err != nil {
		return err
	}

	deployment, err := r.reconcileAdapter(src, ch)
	if err != nil {
		return err
	}

	service, err := r.reconcileService(src, deployment)
	if err != nil {
		return err
	}

	peerAuthentication, err := r.reconcilePeerAuthentication(src, deployment)
	if err != nil {
		return err
	}

	return r.syncStatus(src, ch, deployment, peerAuthentication, service)
}

// httpSourceByKey retrieves a HTTPSource object from a lister by ns/name key.
func httpSourceByKey(key string, l sourceslistersv1alpha1.HTTPSourceLister) (*sourcesv1alpha1.HTTPSource, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "invalid object key"))
	}

	src, err := l.HTTPSources(ns).Get(name)
	switch {
	case apierrors.IsNotFound(err):
		return nil, errors.NewSkippable(pkgerrors.Wrap(err, "object no longer exists"))
	case err != nil:
		return nil, err
	}

	return src, nil
}

// reconcileSink reconciles the state of the HTTP adapter's sink (Channel).
func (r *Reconciler) reconcileSink(src *sourcesv1alpha1.HTTPSource) (*messagingv1alpha1.Channel, error) {
	desiredCh := r.makeChannel(src)

	currentCh, err := r.getOrCreateChannel(src, desiredCh)
	if err != nil {
		return nil, err
	}

	object.ApplyExistingChannelAttributes(currentCh, desiredCh)

	currentCh, err = r.syncChannel(src, currentCh, desiredCh)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Channel")
	}

	return currentCh, nil
}

// reconcileService reconciles the state of the kubernetes Service.
func (r *Reconciler) reconcileService(src *sourcesv1alpha1.HTTPSource, deployment *appsv1.Deployment) (*corev1.Service, error) {
	if deployment == nil {
		r.Logger.Info("Skipping creation of Service as there is no deployment yet")
		return nil, nil
	}
	desiredService := r.makeService(src)
	currentService, err := r.getOrCreateService(src, desiredService)
	if err != nil {
		return nil, err
	}

	object.ApplyExistingServiceAttributes(currentService, desiredService)

	_, err = r.syncService(src, currentService, desiredService)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Service")
	}

	return desiredService, nil
}

// reconcilePeerAuthentication reconciles the state of the PeerAuthentication.
func (r *Reconciler) reconcilePeerAuthentication(src *sourcesv1alpha1.HTTPSource, deployment *appsv1.Deployment) (*securityv1beta1.PeerAuthentication, error) {
	if deployment == nil {
		r.Logger.Info("Skipping creation of Istio PeerAuthentication as there is no deployment yet")
		return nil, nil
	}

	if deployment.Status.AvailableReplicas == 0 {
		r.Logger.Info("Skipping creation of Istio PeerAuthentication as there is no revision yet")
		return nil, nil
	}
	desiredPeerAuthentication := r.makePeerAuthentication(src, deployment)

	currentPeerAuthentication, err := r.getOrCreatePeerAuthentication(src, desiredPeerAuthentication)
	if err != nil {
		return nil, err
	}
	object.ApplyExistingPeerAuthenticationAttributes(currentPeerAuthentication, desiredPeerAuthentication)

	_, err = r.syncPeerAuthentication(src, currentPeerAuthentication, desiredPeerAuthentication)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Istio PeerAuthentication")
	}

	return desiredPeerAuthentication, nil
}

// reconcileAdapter reconciles the state of the HTTP adapter.
func (r *Reconciler) reconcileAdapter(src *sourcesv1alpha1.HTTPSource,
	sink *messagingv1alpha1.Channel) (*appsv1.Deployment, error) {

	sinkURI, err := getSinkURI(sink, r.sinkResolver, src)
	if err != nil {
		// delay reconciliation until the sink becomes ready
		return nil, nil
	}

	desiredDeployment := r.makeDeployment(src, sinkURI, r.adapterLoggingCfg, r.adapterMetricsCfg)

	currentDeployment, err := r.getOrCreateDeployment(src, desiredDeployment)
	if err != nil {
		return nil, err
	}

	object.ApplyExistingDeploymentAttributes(currentDeployment, desiredDeployment)

	currentDeployment, err = r.syncDeployment(src, currentDeployment, desiredDeployment)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Deployment")
	}

	return currentDeployment, nil
}

// getOrCreateDeployment returns the existing Deployment for a given
// HTTPSource, or creates it if it is missing.
func (r *Reconciler) getOrCreateDeployment(src *sourcesv1alpha1.HTTPSource,
	desiredDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {

	deployment, err := r.deploymentLister.Deployments(src.Namespace).Get(src.Name)
	switch {
	case apierrors.IsNotFound(err):
		deployment, err = r.KubeClientSet.AppsV1().Deployments(src.Namespace).Create(desiredDeployment)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Deployment %q", desiredDeployment.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Deployment")
		}
		r.event(src, createReason, "Created Deployment %q", deployment.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Deployment from cache")
	}

	return deployment, nil
}

// getOrCreateChannel returns the existing Channel for a given HTTPSource, or
// creates it if it is missing.
func (r *Reconciler) getOrCreateChannel(src *sourcesv1alpha1.HTTPSource,
	desiredCh *messagingv1alpha1.Channel) (*messagingv1alpha1.Channel, error) {

	ch, err := r.chLister.Channels(src.Namespace).Get(src.Name)
	switch {
	case apierrors.IsNotFound(err):
		ch, err = r.messagingClient.Channels(src.Namespace).Create(desiredCh)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Channel %q", desiredCh.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Channel")
		}
		r.event(src, createReason, "Created Channel %q", ch.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Channel from cache")
	}

	return ch, nil
}

// getOrCreateService returns the existing Service for a Deployment, or
// creates it if it is missing.
func (r *Reconciler) getOrCreateService(src *sourcesv1alpha1.HTTPSource,
	desiredService *corev1.Service) (*corev1.Service, error) {
	service, err := r.serviceLister.Services(src.Namespace).Get(desiredService.Name)
	switch {
	case apierrors.IsNotFound(err):
		service, err = r.KubeClientSet.CoreV1().Services(src.Namespace).Create(desiredService)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Service %q", desiredService.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Service")
		}
		r.event(src, createReason, "Created Service %q", service.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Service from cache")
	}

	return service, nil
}

// getOrCreatePeerAuthentication returns the existing PeerAuthentication for a Replica of a Deployment, or
// creates it if it is missing.
func (r *Reconciler) getOrCreatePeerAuthentication(src *sourcesv1alpha1.HTTPSource,
	desiredPeerAuthentication *securityv1beta1.PeerAuthentication) (*securityv1beta1.PeerAuthentication, error) {
	peerAuthentication, err := r.peerAuthenticationLister.PeerAuthentications(src.Namespace).Get(desiredPeerAuthentication.Name)
	switch {
	case apierrors.IsNotFound(err):
		peerAuthentication, err = r.securityClient.PeerAuthentications(src.Namespace).Create(desiredPeerAuthentication)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Istio PeerAuthentication %q", desiredPeerAuthentication.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Istio PeerAuthentication")
		}
		r.event(src, createReason, "Created Istio PeerAuthentication %q", peerAuthentication.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Istio PeerAuthentication from cache")
	}

	return peerAuthentication, nil
}

// makeDeployment returns the desired Deployment object for a given
// HTTPSource. An optional Deployment can be passed as parameter, in which
// case some of its attributes are used to generate the desired state.
func (r *Reconciler) makeDeployment(src *sourcesv1alpha1.HTTPSource,
	sinkURI, loggingCfg, metricsCfg string) *appsv1.Deployment {

	return object.NewDeployment(src.Namespace, src.Name,
		object.WithReplicas(1),
		object.WithName(adapterContainerName),
		object.WithImage(r.adapterEnvCfg.Image),
		object.WithPort(r.adapterEnvCfg.Port, portName),
		object.WithPort(metricsPort, metricsPortName),
		object.WithEnvVar(eventSourceEnvVar, src.Spec.Source),
		object.WithEnvVar(sinkURIEnvVar, sinkURI),
		object.WithEnvVar(namespaceEnvVar, src.Namespace),
		object.WithEnvVar(metricsConfigEnvVar, metricsCfg),
		object.WithEnvVar(loggingConfigEnvVar, loggingCfg),
		object.WithEnvVar(adapterPortEnvVar, strconv.Itoa(adapterPort)),
		object.WithProbe(adapterHealthEndpoint, adapterPort),
		object.WithControllerRef(src.ToOwner()),
		object.WithLabel(applicationNameLabelKey, src.Name),
		object.WithMatchLabelsSelector(applicationNameLabelKey, src.Name),
		object.WithPodLabel(applicationNameLabelKey, src.Name),
		object.WithPodLabel(applicationLabelKey, src.Name),
		object.WithPodLabel(dashboardLabelKey, dashboardLabelValue),
		object.WithPodLabel(eventSourceLabelKey, eventSourceLabelValue),
	)
}

// makeChannel returns the desired Channel object for a given HTTPSource.
func (r *Reconciler) makeChannel(src *sourcesv1alpha1.HTTPSource) *messagingv1alpha1.Channel {
	return object.NewChannel(src.Namespace, src.Name,
		object.WithControllerRef(src.ToOwner()),
		object.WithLabel(applicationNameLabelKey, src.Name),
	)
}

// makeService returns the desired Service object for a given HTTPSource.
func (r *Reconciler) makeService(src *sourcesv1alpha1.HTTPSource) *corev1.Service {

	return object.NewService(src.Namespace, src.Name,
		object.WithControllerRef(src.ToOwner()),
		object.WithSelector(applicationNameLabelKey, src.Name),
		object.WithServicePort(portName, externalPort, int(r.adapterEnvCfg.Port)),
		object.WithServicePort(metricsPortName, metricsPort, metricsPort),
		object.WithLabel(applicationNameLabelKey, src.Name),
		object.WithLabel(dashboardLabelKey, dashboardLabelValue),
	)
}

// makePeerAuthentication returns the desired PeerAuthentication object for a given Deployment per HTTPSource.
func (r *Reconciler) makePeerAuthentication(src *sourcesv1alpha1.HTTPSource, deployment *appsv1.Deployment) *securityv1beta1.PeerAuthentication {
	// Using the private k8s svc of a deployment which has the metrics ports
	return object.NewPeerAuthentication(src.Namespace, deployment.Name,
		object.WithControllerRef(src.ToOwner()),
		object.WithLabel(applicationNameLabelKey, src.Name),
		object.WithSelectorSpec(map[string]string{applicationLabelKey: src.Name}),
		object.WithPermissiveMode(metricsPort),
	)
}

// syncDeployment synchronizes the desired state of a Deployment against
// its current state in the running cluster.
func (r *Reconciler) syncDeployment(src *sourcesv1alpha1.HTTPSource,
	currentDeployment, desiredDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {

	if object.Semantic.DeepEqual(currentDeployment, desiredDeployment) {
		return currentDeployment, nil
	}

	deployment, err := r.KubeClientSet.AppsV1().Deployments(currentDeployment.Namespace).Update(desiredDeployment)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Deployment %q", desiredDeployment.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Deployment %q", deployment.Name)

	return deployment, nil
}

// syncChannel synchronizes the desired state of a Channel against its current
// state in the running cluster.
func (r *Reconciler) syncChannel(src *sourcesv1alpha1.HTTPSource,
	currentCh, desiredCh *messagingv1alpha1.Channel) (*messagingv1alpha1.Channel, error) {

	if object.Semantic.DeepEqual(currentCh, desiredCh) {
		return currentCh, nil
	}

	ch, err := r.messagingClient.Channels(currentCh.Namespace).Update(desiredCh)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Channel %q", desiredCh.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Channel %q", ch.Name)

	return ch, nil
}

// syncService synchronizes the desired state of a Service against its current
// state in the running cluster.
func (r *Reconciler) syncService(src *sourcesv1alpha1.HTTPSource,
	currentSvc, desiredSvc *corev1.Service) (*corev1.Service, error) {

	if object.Semantic.DeepEqual(currentSvc, desiredSvc) {
		return currentSvc, nil
	}

	svc, err := r.KubeClientSet.CoreV1().Services(currentSvc.Namespace).Update(desiredSvc)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Service %q", desiredSvc.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Service %q", svc.Name)

	return svc, nil
}

// syncPeerAuthentication synchronizes the desired state of a PeerAuthentication against its current
// state in the running cluster.
func (r *Reconciler) syncPeerAuthentication(src *sourcesv1alpha1.HTTPSource, currentPeerAuthentication, desiredPeerAuthentication *securityv1beta1.PeerAuthentication) (*securityv1beta1.PeerAuthentication, error) {

	if object.Semantic.DeepEqual(currentPeerAuthentication, desiredPeerAuthentication) {
		return currentPeerAuthentication, nil
	}

	peerAuthentication, err := r.securityClient.PeerAuthentications(currentPeerAuthentication.Namespace).Update(desiredPeerAuthentication)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Istio PeerAuthentication %q", desiredPeerAuthentication.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Istio PeerAuthentication %q", peerAuthentication.Name)

	return peerAuthentication, nil
}

// syncStatus ensures the status of a given HTTPSource is up-to-date.
func (r *Reconciler) syncStatus(src *sourcesv1alpha1.HTTPSource,
	ch *messagingv1alpha1.Channel, deployment *appsv1.Deployment, peerAuthentication *securityv1beta1.PeerAuthentication, service *corev1.Service) error {

	currentStatus := &src.Status
	expectedStatus := r.computeStatus(src, ch, deployment, peerAuthentication, service)

	if reflect.DeepEqual(currentStatus, expectedStatus) {
		return nil
	}

	src = &sourcesv1alpha1.HTTPSource{
		ObjectMeta: src.ObjectMeta,
		Status:     *expectedStatus,
		// sending the Spec in a status update is optional, however
		// fake ClientSets apply the same UpdateActionImpl action to
		// the object tracker regardless of the subresource
		Spec: src.Spec,
	}

	_, err := r.sourcesClient.HTTPSources(src.Namespace).UpdateStatus(src)
	return err
}

// computeStatus returns the expected status of a given HTTPSource.
func (r *Reconciler) computeStatus(src *sourcesv1alpha1.HTTPSource, ch *messagingv1alpha1.Channel,
	deployment *appsv1.Deployment, peerAuthentication *securityv1beta1.PeerAuthentication, service *corev1.Service) *sourcesv1alpha1.HTTPSourceStatus {

	status := src.Status.DeepCopy()
	status.InitializeConditions()

	sinkURI, err := r.sinkResolver.URIFromDestination(channelAsDestination(ch), src)
	if err != nil {
		status.MarkNoSink()
		return status
	}
	status.MarkSink(sinkURI)
	status.MarkPeerAuthenticationCreated(peerAuthentication)
	status.MarkServiceCreated(service)

	if deployment != nil {
		status.PropagateDeploymentReady(deployment)
	}

	return status
}

// getSinkURI returns the URI of a Channel. An error is returned if the URI
// can't be determined which, can also happen when the Channel is not ready.
func getSinkURI(sink *messagingv1alpha1.Channel, r *resolver.URIResolver, parent interface{}) (string, error) {
	sinkURI, err := r.URIFromDestination(channelAsDestination(sink), parent)
	if err != nil {
		return "", pkgerrors.Wrap(err, "failed to read sink URI")
	}
	return sinkURI, nil
}

// channelAsDestination returns a Destination representation of the given
// Channel.
func channelAsDestination(ch *messagingv1alpha1.Channel) apisv1beta1.Destination {
	gvk := ch.GetGroupVersionKind()

	return apisv1beta1.Destination{
		Ref: &corev1.ObjectReference{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Namespace:  ch.Namespace,
			Name:       ch.Name,
		},
	}
}
