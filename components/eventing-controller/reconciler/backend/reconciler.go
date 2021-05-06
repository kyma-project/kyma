package backend

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/go-logr/logr"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// This probably deserves a better name...
	BEBBackendSecretLabelKey   = "kyma-project.io/eventing-backend"
	BEBBackendSecretLabelValue = "beb"
	DefaultEventingBackendName = "eventing-backend"
	// TODO: where to get this namespace
	DefaultEventingBackendNamespace = "kyma-system"

	PublisherNamespace       = "kyma-system"
	PublisherName            = "eventing-publisher-proxy"
	// TODO: Use a more generic name, once added to the helm chart!
	ServiceAccountName       = "eventing-event-publisher-nats"
	BackendCRLabelKey        = "kyma-project.io/eventing"
	BackendCRLabelValue      = "backend"
	AppLabelKey              = "app.kubernetes.io/name"
	AppLabelValue            = PublisherName
	InstanceLabelKey         = "app.kubernetes.io/instance"
	InstanceLabelValue       = "eventing"
	DashboardLabelKey        = "kyma-project.io/dashboard"
	DashboardLabelValue      = "eventing"
	PublisherReplicas        = 1
	PublisherImage           = "eu.gcr.io/kyma-project/event-publisher-proxy:88360eed"
	PublisherPortName        = "http"
	PublisherPortNum         = int32(8080)
	PublisherMetricsPortName = "http-metrics"
	PublisherMetricsPortNum  = int32(9090)

	LivenessInitialDelaySecs = int32(5)
	LivenessTimeoutSecs      = int32(1)
	LivenessPeriodSecs       = int32(2)
	BEBNamespacePrefix       = "/"
)

var (
	TerminationGracePeriodSeconds = int64(30)
)

type BackendReconciler struct {
	client.Client
	cache.Cache
	Log logr.Logger
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends/status,verbs=get;update;patch

func (r *BackendReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var secretList v1.SecretList

	if err := r.Cache.List(ctx, &secretList, client.MatchingLabels{
		BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
	}); err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Found secrets with label", "count", len(secretList.Items))

	if len(secretList.Items) > 1 {
		// Break the system
		// For now ignore...
		return ctrl.Result{}, nil
	}

	backendType := eventingv1alpha1.NatsBackendType
	// If the label is removed what to do? first check if the removed secret is mentioned in the backend CR?
	// Don't we need a finalizer for that?
	if len(secretList.Items) == 1 {
		r.Log.Info("***** Going with the BEB flow ***")
		backendType = eventingv1alpha1.BebBackendType
		// CreateOrUpdate CR with BEB
		_, err := r.CreateOrUpdateBackendCR(ctx, backendType)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Stop subscription controller (Radu/Frank)
		// Start the other subscription controller (Radu/Frank)
		// CreateOrUpdate deployment for publisher proxy
		publisher, err := r.CreateOrUpdatePublisherProxy(ctx, backendType)
		if err != nil {
			// Update status if bad
			r.Log.Error(err, "failed to create or update publisher proxy", "backend", backendType)
			return ctrl.Result{}, err
		}
		// CreateOrUpdate status of the CR
		err = r.UpdateBackendStatus(ctx, backendType, publisher)
		if err != nil {
			r.Log.Error(err, "failed to create or update backend status", "backend", backendType)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// NATS flow
	// CreateOrUpdate CR with NATS
	r.Log.Info("***** Going with the NATS flow ***")
	_, err := r.CreateOrUpdateBackendCR(ctx, backendType)
	if err != nil {
		// Update status if bad
		return ctrl.Result{}, err
	}
	r.Log.Info("Created/updated backend CR")
	// Stop subscription controller (Radu/Frank)
	// Start the other subscription controller (Radu/Frank)
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
	err = r.UpdateBackendStatus(ctx, backendType, publisher)
	return ctrl.Result{}, err
}

func (r *BackendReconciler) UpdateBackendStatus(ctx context.Context, backendType eventingv1alpha1.BackendType, publisher *appsv1.Deployment) error {
	var backend eventingv1alpha1.EventingBackend
	// TODO: cache?
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: DefaultEventingBackendNamespace,
		Name:      DefaultEventingBackendName,
	}, &backend); err != nil {
		r.Log.Info("cannot get backend CR to update")
		return err
	}

	// TODO: in case a publisher already exists, to make sure during the switch the status of publisherReady is false,
	//  do we need to make sure first we take the existing publisher down, then patch the existing deployment? Otherwise,
	//  during the transition from nats<->beb, there are two replicas available.
	publisherReady := *publisher.Spec.Replicas == publisher.Status.ReadyReplicas

	backend.Status = eventingv1alpha1.EventingBackendStatus{
		Backend:         backendType,
		ControllerReady: boolPtr(false),
		EventingReady:   boolPtr(false),
		PublisherReady:  boolPtr(publisherReady),
	}
	r.Log.Info("Updating backend CR status")
	if err := r.Update(ctx, &backend); err != nil {
		r.Log.Error(err, "error updating EventingBackend CR")
		return err
	}
	return nil
}

func (r *BackendReconciler) CreateOrUpdatePublisherProxy(ctx context.Context, backend eventingv1alpha1.BackendType) (*appsv1.Deployment, error) {
	publisherNamespacedName := types.NamespacedName{
		Namespace: PublisherNamespace,
		Name:      PublisherName,
	}
	var currentPublisher appsv1.Deployment
	var desiredPublisher *appsv1.Deployment

	switch backend {
	case eventingv1alpha1.NatsBackendType:
		desiredPublisher = newNATSPublisherDeployment()
	case eventingv1alpha1.BebBackendType:
		desiredPublisher = newBEBPublisherDeployment()
	default:
		return nil, fmt.Errorf("unknown eventing backend type %q", backend)
	}

	err := r.Cache.Get(ctx, publisherNamespacedName, &currentPublisher)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create
			r.Log.Info("creating publisher proxy")
			return desiredPublisher, r.Create(ctx, desiredPublisher)
		}
		return nil, err
	}

	desiredPublisher.ResourceVersion = currentPublisher.ResourceVersion
	if publisherProxyDeploymentEqual(&currentPublisher, desiredPublisher) {
		return &currentPublisher, nil
	}

	// Update if necessary
	if err := r.Update(ctx, desiredPublisher); err != nil {
		r.Log.Error(err, "Cannot update publisher proxy")
		return nil, err
	}
	err = r.UpdateBackendStatus(ctx, backend, desiredPublisher)
	if err != nil {
		r.Log.Error(err, "Cannot update the status for EventingBackend CR")
		return nil, err
	}

	return desiredPublisher, nil
}

func newBEBPublisherDeployment() *appsv1.Deployment {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		InstanceLabelKey:  InstanceLabelValue,
		DashboardLabelKey: DashboardLabelValue,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PublisherName,
			Namespace: PublisherNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: intPtr(PublisherReplicas),
			Selector: metav1.SetAsLabelSelector(labels),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   PublisherName,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  PublisherName,
							Image: PublisherImage,
							Ports: []v1.ContainerPort{
								{
									Name:          PublisherPortName,
									ContainerPort: PublisherPortNum,
								},
								{
									Name:          PublisherMetricsPortName,
									ContainerPort: PublisherMetricsPortNum,
								},
							},
							Env: []v1.EnvVar{
								{Name: "BACKEND", Value: "beb"},
								{Name: "PORT", Value: strconv.Itoa(int(PublisherPortNum))},
								{Name: "REQUEST_TIMEOUT", Value: "5s"},
								{Name: "EVENT_TYPE_PREFIX", Value: "sap.kyma.custom"},
								{
									Name: "CLIENT_ID",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "client-id",
										}},
								},
								{
									Name: "CLIENT_SECRET",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "client-secret",
										}},
								},
								{
									Name: "TOKEN_ENDPOINT",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "token-endpoint",
										}},
								},
								{
									Name: "EMS_PUBLISH_URL",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "ems-publish-url",
										}},
								},
								{
									Name: "BEB_NAMESPACE_VALUE",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "beb-namespace",
										}},
								},
								{
									Name:  "BEB_NAMESPACE",
									Value: fmt.Sprintf("%s$(BEB_NAMESPACE_VALUE)", BEBNamespacePrefix),
								},
								{
									Name: "WEBHOOK_CLIENT_ID",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing-controller-beb-oauth2"},
											Key:                  "client_id",
										}},
								},
								{
									Name: "WEBHOOK_CLIENT_SECRET",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing-controller-beb-oauth2"},
											Key:                  "client_secret",
										}},
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: LivenessInitialDelaySecs,
								TimeoutSeconds:      LivenessTimeoutSecs,
								PeriodSeconds:       LivenessPeriodSecs,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								FailureThreshold: 3,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: &v1.SecurityContext{
								Privileged:               boolPtr(false),
								AllowPrivilegeEscalation: boolPtr(false),
							},
						},
					},
					RestartPolicy:                 v1.RestartPolicyAlways,
					ServiceAccountName:            ServiceAccountName,
					TerminationGracePeriodSeconds: &TerminationGracePeriodSeconds,
				},
			},
		},
	}
}

func newNATSPublisherDeployment() *appsv1.Deployment {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		InstanceLabelKey:  InstanceLabelValue,
		DashboardLabelKey: DashboardLabelValue,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PublisherName,
			Namespace: PublisherNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: intPtr(PublisherReplicas),
			Selector: metav1.SetAsLabelSelector(labels),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   PublisherName,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  PublisherName,
							Image: PublisherImage,
							Ports: []v1.ContainerPort{
								{
									Name:          PublisherPortName,
									ContainerPort: PublisherPortNum,
								},
								{
									Name:          PublisherMetricsPortName,
									ContainerPort: PublisherMetricsPortNum,
								},
							},
							Env: []v1.EnvVar{
								{Name: "BACKEND", Value: "nats"},
								{Name: "PORT", Value: strconv.Itoa(int(PublisherPortNum))},
								{Name: "NATS_URL", Value: "eventing-nats.kyma-system.svc.cluster.local"},
								{Name: "REQUEST_TIMEOUT", Value: "5s"},
								{Name: "LEGACY_NAMESPACE", Value: "kyma"},
								{Name: "LEGACY_EVENT_TYPE_PREFIX", Value: "sap.kyma.custom"},
								{Name: "EVENT_TYPE_PREFIX", Value: "sap.kyma.custom"},
								{
									Name: "CLIENT_ID",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "client-id",
										}},
								},
								{
									Name: "CLIENT_SECRET",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "eventing"},
											Key:                  "client-secret",
										}},
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: LivenessInitialDelaySecs,
								TimeoutSeconds:      LivenessTimeoutSecs,
								PeriodSeconds:       LivenessPeriodSecs,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								FailureThreshold: 3,
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: &v1.SecurityContext{
								Privileged:               boolPtr(false),
								AllowPrivilegeEscalation: boolPtr(false),
							},
						},
					},
					RestartPolicy:                 v1.RestartPolicyAlways,
					ServiceAccountName:            ServiceAccountName,
					TerminationGracePeriodSeconds: &TerminationGracePeriodSeconds,
				},
			},
			Strategy: appsv1.DeploymentStrategy{},
		},
		Status: appsv1.DeploymentStatus{},
	}
}

func (r *BackendReconciler) CreateOrUpdateBackendCR(ctx context.Context, backend eventingv1alpha1.BackendType) (*eventingv1alpha1.EventingBackend, error) {
	var currentBackend eventingv1alpha1.EventingBackend
	defaultEventingBackend := types.NamespacedName{
		Namespace: DefaultEventingBackendNamespace,
		Name:      DefaultEventingBackendName,
	}

	labels := map[string]string{
		BackendCRLabelKey: BackendCRLabelValue,
	}
	desiredBackend := eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultEventingBackendName,
			Namespace: DefaultEventingBackendNamespace,
			Labels:    labels,
		},
		Spec: eventingv1alpha1.EventingBackendSpec{},
		Status: eventingv1alpha1.EventingBackendStatus{
			Backend:            backend,
			EventingReady:      boolPtr(false),
			ControllerReady:    boolPtr(false),
			PublisherReady:     boolPtr(false),
			BebSecretName:      "",
			BebSecretNamespace: "",
		},
	}

	err := r.Cache.Get(ctx, defaultEventingBackend, &currentBackend)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("trying to create backend CR...")
			if err := r.Create(ctx, &desiredBackend); err != nil {
				r.Log.Error(err, "Cannot create an EventingBackend CR")
				return nil, err
			}
			r.Log.Info("created backend CR")
			return &desiredBackend, nil
		}
		r.Log.Info("error is not NotFound!", "err", err)
		return nil, err
	}
	r.Log.Info("Found existing backend CR")

	desiredBackend.ResourceVersion = currentBackend.ResourceVersion
	if eventingBackendEqual(&currentBackend, &desiredBackend) {
		r.Log.Info("No need to update existing backend CR")
		return &currentBackend, nil
	}
	r.Log.Info("Update exiting backend CR")
	if err := r.Update(ctx, &desiredBackend); err != nil {
		r.Log.Error(err, "Cannot update an EventingBackend CR")
		return nil, err
	}

	return &desiredBackend, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int32) *int32 {
	return &i
}

func (r *BackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Secret{}).Complete(r)
}
