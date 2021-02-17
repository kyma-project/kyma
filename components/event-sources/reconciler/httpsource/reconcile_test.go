package httpsource

import (
	"context"
	"strconv"
	"testing"

	pkgerrors "github.com/pkg/errors"
	istiov1beta1apis "istio.io/api/type/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8stesting "k8s.io/client-go/testing"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/client/injection/ducks/duck/v1/addressable"
	_ "knative.dev/pkg/client/injection/ducks/duck/v1/addressable/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/ptr"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"

	securityv1beta1apis "istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	fakesourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	fakeistioclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/client/fake"
	. "github.com/kyma-project/kyma/components/event-sources/reconciler/testing"
)

const (
	tNs                 = "testns"
	tName               = "test"
	tUID                = types.UID("00000000-0000-0000-0000-000000000000")
	tImg                = "sources.kyma-project.io/http:latest"
	tExternalPort       = 80
	tPort               = 8080
	tSinkURI            = "http://" + tName + "-kn-channel." + tNs + ".svc.cluster.local"
	tSource             = "varkes"
	tPeerAuthentication = "test"
	tRevisionSvc        = "test"
	tTargetPort         = metricsPort
	tMetricsDomain      = "testing"
)

var tOwnerRef = metav1.OwnerReference{
	APIVersion:         sourcesv1alpha1.HTTPSourceGVK().GroupVersion().String(),
	Kind:               sourcesv1alpha1.HTTPSourceGVK().Kind,
	Name:               tName,
	UID:                tUID,
	Controller:         ptr.Bool(true),
	BlockOwnerDeletion: ptr.Bool(true),
}

var tChLabels = map[string]string{
	applicationNameLabelKey: tName,
}

var (
	tMetricsData = map[string]string{
		"metrics.backend": "prometheus",
	}

	tLoggingData = map[string]string{
		"zap-logger-config": `{"level": "info"}`,
	}
)

var tEnvVars = []corev1.EnvVar{
	{
		Name:  eventSourceEnvVar,
		Value: tSource,
	}, {
		Name:  sinkURIEnvVar,
		Value: tSinkURI,
	}, {
		Name:  namespaceEnvVar,
		Value: tNs,
	}, {
		Name: metricsConfigEnvVar,
		Value: `{"Domain":"` + tMetricsDomain + `",` +
			`"Component":"` + component + `",` +
			`"PrometheusPort":` + strconv.Itoa(adapterMetricsPort) + `,` +
			`"ConfigMap":{"metrics.backend":"prometheus"}}`,
	}, {
		Name:  loggingConfigEnvVar,
		Value: `{"zap-logger-config":"{\"level\": \"info\"}"}`,
	}, {
		Name:  adapterPortEnvVar,
		Value: strconv.Itoa(adapterPort),
	},
}

func TestReconcile(t *testing.T) {
	testCases := rt.TableTest{
		/* Error handling */

		{
			Name:    "Source was deleted",
			Key:     tNs + "/" + tName,
			Objects: nil,
			WantErr: false,
		},
		{
			Name:    "Invalid object key",
			Key:     tNs + "/" + tName + "/invalid",
			WantErr: true,
		},

		/* Service synchronization */

		{
			Name: "Initial source creation",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(),
			},
			WantCreates: []runtime.Object{
				newChannelNotReady(),
				// no Service gets created until the Channel
				// becomes ready
			},
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSource(WithoutSink), // "Deployed" condition remains Unknown
			}},
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(createReason), "Created Channel %q", tName),
			},
		},
		{
			Name: "Everything up-to-date",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithSink, WithPeerAuthentication, WithService),
				newDeploymentAvailable(),
				newChannelReady(),
				newPeerAuthenticationWithSpec(),
				newService(),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
		},
		{
			Name: "Adapter Service spec does not match expectation",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithSink, WithPeerAuthentication, WithService),
				func() *appsv1.Deployment {
					deploy := newDeploymentAvailable()
					deploy.Labels["some-label"] = "unexpected"
					return deploy
				}(),
				newChannelReady(),
				newPeerAuthenticationWithSpec(),
				newService(),
			},
			WantCreates: nil,
			WantUpdates: []k8stesting.UpdateActionImpl{{
				Object: newDeploymentAvailable(),
			}},
			WantStatusUpdates: nil,
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(updateReason), "Updated Deployment %q", tName),
			},
		},

		/* Channel synchronization */

		{
			Name: "Channel missing",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithoutSink),
				newDeploymentAvailable(),
			},
			WantCreates: []runtime.Object{
				newChannelNotReady(),
			},
			WantUpdates:       nil,
			WantStatusUpdates: nil,
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(createReason), "Created Channel %q", tName),
			},
		},
		{
			Name: "Channel spec does not match expectation",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithSink, WithPeerAuthentication, WithService),
				newDeploymentAvailable(),
				func() *messagingv1alpha1.Channel {
					ch := newChannelReady()
					ch.Labels["some-label"] = "unexpected"
					return ch
				}(),
				newService(),
				newPeerAuthenticationWithSpec(),
			},
			WantCreates: nil,
			WantUpdates: []k8stesting.UpdateActionImpl{{
				Object: newChannelReady(),
			}},
			WantStatusUpdates: nil,
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(updateReason), "Updated Channel %q", tName),
			},
		},

		/* PeerAuthentication synchronization */

		{
			Name: "PeerAuthentication missing when deployment not available",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(NotDeployed, WithSink, WithoutPeerAuthentication, WithService),
				newDeploymentNotAvailable(),
				newChannelReady(),
				newService(),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
			WantEvents:        nil,
		},
		{
			Name: "PeerAuthentication created",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithSink, WithPeerAuthentication, WithService),
				newDeploymentAvailable(),
				newChannelReady(),
				newService(),
			},
			WantCreates: []runtime.Object{
				newPeerAuthenticationWithSpec(),
			},
			WantUpdates:       nil,
			WantStatusUpdates: nil,
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(createReason), "Created Istio PeerAuthentication %q", tPeerAuthentication),
			},
		},
		{
			Name: "Adapter deployment in progress",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(NotDeployed, WithSink, WithoutPeerAuthentication, WithService),
				newDeploymentNotAvailable(),
				newChannelReady(),
				newService(),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
			WantEvents:        nil,
		},
		{
			Name: "Adapter Service became ready",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(NotDeployed, WithSink, WithService, WithPeerAuthentication),
				newDeploymentAvailable(),
				newChannelReady(),
				newPeerAuthenticationWithSpec(),
				newService(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSource(Deployed, WithSink, WithService, WithPeerAuthentication),
			}},
			WantEvents: nil,
		},
		{
			Name: "Adapter Service became not ready",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithSink, WithService, WithoutPeerAuthentication),
				newDeploymentNotAvailable(),
				newChannelReady(),
				newService(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSource(NotDeployed, WithSink, WithService, WithoutPeerAuthentication),
			}},
			WantEvents: nil,
		},
		{
			Name: "Channel becomes available",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithoutSink, WithoutPeerAuthentication, WithService),
				newDeploymentAvailable(),
				newChannelReady(),
				newService(),
			},
			WantCreates: []runtime.Object{
				newPeerAuthenticationWithSpec(),
			},
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSource(Deployed, WithSink, WithPeerAuthentication, WithService),
			}},
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(createReason), "Created Istio PeerAuthentication %q", tPeerAuthentication),
			},
		},
		{
			Name: "Channel becomes unavailable",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSource(Deployed, WithSink, WithService),
				newDeploymentAvailable(),
				newChannelNotReady(),
				newService(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSource(Deployed, WithService, WithoutSink), // previous Deployed status remains
			}},
			WantEvents: nil,
		},
	}

	var ctor Ctor = func(t *testing.T, ctx context.Context, ls *Listers) controller.Reconciler {
		ctx = addressable.WithDuck(ctx)
		defer SetEnvVar(t, metrics.DomainEnv, tMetricsDomain)()

		cmw := configmap.NewStaticWatcher(
			NewConfigMap("", metrics.ConfigMapName(), WithData(tMetricsData)),
			NewConfigMap("", logging.ConfigMapName(), WithData(tLoggingData)),
		)

		rb := reconciler.NewBase(ctx, controllerAgentName, cmw)
		r := &Reconciler{
			Base: rb,
			adapterEnvCfg: &httpAdapterEnvConfig{
				Image: tImg,
				Port:  tPort,
			},
			httpsourceLister:         ls.GetHTTPSourceLister(),
			deploymentLister:         ls.GetDeploymentLister(),
			chLister:                 ls.GetChannelLister(),
			peerAuthenticationLister: ls.GetPeerAuthenticationLister(),
			serviceLister:            ls.GetServiceLister(),
			sourcesClient:            fakesourcesclient.Get(ctx).SourcesV1alpha1(),
			messagingClient:          rb.EventingClientSet.MessagingV1alpha1(),
			sinkResolver:             resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
			securityClient:           fakeistioclient.Get(ctx).SecurityV1beta1(),
		}

		cmw.Watch(metrics.ConfigMapName(), r.updateAdapterMetricsConfig)
		cmw.Watch(logging.ConfigMapName(), r.updateAdapterLoggingConfig)

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}

type SourceOption func(*sourcesv1alpha1.HTTPSource)

// newSource returns a test HTTPSource object with pre-filled metadata.
func newSource(opts ...SourceOption) *sourcesv1alpha1.HTTPSource {
	src := &sourcesv1alpha1.HTTPSource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			UID:       tUID,
		},
		Spec: sourcesv1alpha1.HTTPSourceSpec{
			Source: tSource,
		},
	}

	src.Status.InitializeConditions()

	for _, opt := range opts {
		opt(src)
	}

	return src
}

func WithoutSink(src *sourcesv1alpha1.HTTPSource) {
	src.Status.MarkNoSink()
}

func WithSink(src *sourcesv1alpha1.HTTPSource) {
	src.Status.MarkSink(tSinkURI)
}

func WithService(src *sourcesv1alpha1.HTTPSource) {
	src.Status.MarkServiceCreated(newService())
}

func WithoutService(src *sourcesv1alpha1.HTTPSource) {
	src.Status.MarkServiceCreated(nil)
}

func WithPeerAuthentication(src *sourcesv1alpha1.HTTPSource) {
	src.Status.MarkPeerAuthenticationCreated(newPeerAuthenticationWithSpec())
}

func WithoutPeerAuthentication(src *sourcesv1alpha1.HTTPSource) {
	src.Status.MarkPeerAuthenticationCreated(nil)
}

func Deployed(src *sourcesv1alpha1.HTTPSource) {
	src.Status.PropagateDeploymentReady(newDeploymentAvailable())
}

func NotDeployed(src *sourcesv1alpha1.HTTPSource) {
	src.Status.PropagateDeploymentReady(newDeploymentNotAvailable())
}

// newChannel returns a test Channel object with pre-filled metadata.
func newChannel() *messagingv1alpha1.Channel {
	lbls := make(map[string]string, len(tChLabels))
	for k, v := range tChLabels {
		lbls[k] = v
	}

	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			Labels:          lbls,
			OwnerReferences: []metav1.OwnerReference{tOwnerRef},
		},
	}
}

// newPeerAuthentication returns a test PeerAuthentication object with pre-filled metadata
func newPeerAuthentication() *securityv1beta1.PeerAuthentication {
	lbls := make(map[string]string, len(tChLabels))
	for k, v := range tChLabels {
		lbls[k] = v
	}

	return &securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tPeerAuthentication,
			Labels:          lbls,
			OwnerReferences: []metav1.OwnerReference{tOwnerRef},
		},
	}
}

// newPeerAuthentication returns a test PeerAuthentication object with Spec
func newPeerAuthenticationWithSpec() *securityv1beta1.PeerAuthentication {
	peerAuthentication := newPeerAuthentication()
	peerAuthentication.Spec.PortLevelMtls = map[uint32]*securityv1beta1apis.PeerAuthentication_MutualTLS{
		tTargetPort: {
			Mode: securityv1beta1apis.PeerAuthentication_MutualTLS_PERMISSIVE,
		},
	}
	peerAuthentication.Spec.Selector = &istiov1beta1apis.WorkloadSelector{
		MatchLabels: map[string]string{
			applicationLabelKey: tName,
		},
	}
	return peerAuthentication
}

// addressable
func newChannelReady() *messagingv1alpha1.Channel {
	ch := newChannel()

	parsedURI, err := apis.ParseURL(tSinkURI)
	if err != nil {
		panic(pkgerrors.Wrap(err, "parsing Channel URL"))
	}

	ch.Status.Address = &duckv1alpha1.Addressable{
		Addressable: duckv1beta1.Addressable{
			URL: parsedURI,
		},
	}

	return ch
}

// not addressable
func newChannelNotReady() *messagingv1alpha1.Channel {
	ch := newChannel()
	ch.Status = messagingv1alpha1.ChannelStatus{}
	return ch
}

// newDeployment returns a test Service object with pre-filled metadata.
func newDeployment() *appsv1.Deployment {
	var replicas int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			Labels: map[string]string{
				applicationNameLabelKey: tName,
			},
			OwnerReferences: []metav1.OwnerReference{tOwnerRef},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{applicationNameLabelKey: tName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						dashboardLabelKey:       dashboardLabelValue,
						eventSourceLabelKey:     eventSourceLabelValue,
						applicationNameLabelKey: tName,
						applicationLabelKey:     tName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: tImg,
						Ports: []corev1.ContainerPort{{
							Name:          portName,
							ContainerPort: tPort,
						},
							{
								Name:          metricsPortName,
								ContainerPort: metricsPort,
							},
						},
						Name: adapterContainerName,
						Env:  tEnvVars,
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: adapterHealthEndpoint,
									Port: intstr.FromInt(adapterPort),
								},
							},
						},
					},
					},
				},
			},
		},
	}
}

// Available: True
func newDeploymentAvailable() *appsv1.Deployment {
	deploy := newDeployment()
	deploy.Status.AvailableReplicas = 1
	return deploy
}

// Available: False
func newDeploymentNotAvailable() *appsv1.Deployment {
	deploy := newDeployment()
	deploy.Status.AvailableReplicas = 0
	return deploy
}

func newService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			Labels: map[string]string{
				dashboardLabelKey:       dashboardLabelValue,
				applicationNameLabelKey: tName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       portName,
					Port:       tExternalPort,
					TargetPort: intstr.FromInt(tPort),
				},
				{
					Name:       metricsPortName,
					Port:       metricsPort,
					TargetPort: intstr.FromInt(metricsPort),
				},
			},
			Selector: tChLabels,
		},
	}

}
