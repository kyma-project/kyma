/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpsource

import (
	"context"
	"knative.dev/serving/pkg/apis/autoscaling"
	"strconv"
	"testing"

	pkgerrors "github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/ptr"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	fakeservingclient "knative.dev/serving/pkg/client/injection/client/fake"
	routeconfig "knative.dev/serving/pkg/reconciler/route/config"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	fakesourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-sources/reconciler/testing"
)

const (
	tNs      = "testns"
	tName    = "test"
	tUID     = types.UID("00000000-0000-0000-0000-000000000000")
	tImg     = "sources.kyma-project.io/http:latest"
	tPort    = 8080
	tSinkURI = "http://" + tName + "-kn-channel." + tNs + ".svc.cluster.local"
	tSource  = "varkes"

	tMetricsDomain = "testing"
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
				Object: newSourceWithoutSink(), // "Deployed" condition remains Unknown
			}},
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(createReason), "Created Channel %q", tName),
			},
		},
		{
			Name: "Everything up-to-date",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceDeployedWithSink(),
				newServiceReady(),
				newChannelReady(),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
		},
		{
			Name: "Adapter Service spec does not match expectation",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceDeployedWithSink(),
				func() *servingv1alpha1.Service {
					svc := newServiceReady()
					svc.Labels["some-label"] = "unexpected"
					return svc
				}(),
				newChannelReady(),
			},
			WantCreates: nil,
			WantUpdates: []k8stesting.UpdateActionImpl{{
				Object: newServiceReady(),
			}},
			WantStatusUpdates: nil,
			WantEvents: []string{
				rt.Eventf(corev1.EventTypeNormal, string(updateReason), "Updated Knative Service %q", tName),
			},
		},

		/* Channel synchronization */

		{
			Name: "Channel missing",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceDeployedWithoutSink(),
				newServiceReady(),
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
				newSourceDeployedWithSink(),
				newServiceReady(),
				func() *messagingv1alpha1.Channel {
					ch := newChannelReady()
					ch.Labels["some-label"] = "unexpected"
					return ch
				}(),
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

		/* Status updates */

		{
			Name: "Adapter Service deployment in progress",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceNotDeployedWithSink(),
				newServiceNotReady(),
				newChannelReady(),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
		},
		{
			Name: "Adapter Service became ready",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceNotDeployedWithSink(),
				newServiceReady(),
				newChannelReady(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSourceDeployedWithSink(),
			}},
		},
		{
			Name: "Adapter Service became not ready",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceDeployedWithSink(),
				newServiceNotReady(),
				newChannelReady(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSourceNotDeployedWithSink(),
			}},
		},
		{
			Name: "Channel becomes available",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceDeployedWithoutSink(),
				newServiceReady(),
				newChannelReady(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSourceDeployedWithSink(),
			}},
		},
		{
			Name: "Channel becomes unavailable",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				newSourceDeployedWithSink(),
				newServiceNotReady(),
				newChannelNotReady(),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: newSourceDeployedWithoutSink(), // previous Deployed status remains
			}},
		},
	}

	var ctor Ctor = func(t *testing.T, ctx context.Context, ls *Listers) controller.Reconciler {
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
			httpsourceLister: ls.GetHTTPSourceLister(),
			ksvcLister:       ls.GetServiceLister(),
			chLister:         ls.GetChannelLister(),
			sourcesClient:    fakesourcesclient.Get(ctx).SourcesV1alpha1(),
			servingClient:    fakeservingclient.Get(ctx).ServingV1alpha1(),
			messagingClient:  rb.EventingClientSet.MessagingV1alpha1(),
			sinkResolver:     resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
		}

		cmw.Watch(metrics.ConfigMapName(), r.updateAdapterMetricsConfig)
		cmw.Watch(logging.ConfigMapName(), r.updateAdapterLoggingConfig)

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}

// newSource returns a test HTTPSource object with pre-filled metadata.
func newSource() *sourcesv1alpha1.HTTPSource {
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

	return src
}

// Deployed: Unknown, SinkProvided: False
func newSourceWithoutSink() *sourcesv1alpha1.HTTPSource {
	src := newSource()
	src.Status.MarkNoSink()
	return src
}

// Deployed: True, SinkProvided: True
func newSourceDeployedWithSink() *sourcesv1alpha1.HTTPSource {
	src := newSource()
	src.Status.PropagateServiceReady(newServiceReady())
	src.Status.MarkSink(tSinkURI)
	return src
}

// Deployed: True, SinkProvided: False
func newSourceDeployedWithoutSink() *sourcesv1alpha1.HTTPSource {
	src := newSource()
	src.Status.PropagateServiceReady(newServiceReady())
	src.Status.MarkNoSink()
	return src
}

// Deployed: False, SinkProvided: True
func newSourceNotDeployedWithSink() *sourcesv1alpha1.HTTPSource {
	src := newSource()
	src.Status.PropagateServiceReady(newServiceNotReady())
	src.Status.MarkSink(tSinkURI)
	return src
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

// newService returns a test Service object with pre-filled metadata.
func newService() *servingv1alpha1.Service {
	return &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			Labels: map[string]string{
				routeconfig.VisibilityLabelKey: routeconfig.VisibilityClusterLocal,
			},
			OwnerReferences: []metav1.OwnerReference{tOwnerRef},
		},
		Spec: servingv1alpha1.ServiceSpec{
			ConfigurationSpec: servingv1alpha1.ConfigurationSpec{
				Template: &servingv1alpha1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							autoscaling.MinScaleAnnotationKey: "1",
						},
					},
					Spec: servingv1alpha1.RevisionSpec{
						RevisionSpec: servingv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: tImg,
									Ports: []corev1.ContainerPort{{
										ContainerPort: tPort,
									}},
									Env: tEnvVars,
									ReadinessProbe: &corev1.Probe{
										Handler: corev1.Handler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: adapterHealthEndpoint,
											},
										},
									},
								}},
							},
						},
					},
				},
			},
		},
	}
}

// Ready: True
func newServiceReady() *servingv1alpha1.Service {
	svc := newService()
	svc.Status.SetConditions(apis.Conditions{{
		Type:   apis.ConditionReady,
		Status: corev1.ConditionTrue,
	}})
	return svc
}

// Ready: False
func newServiceNotReady() *servingv1alpha1.Service {
	svc := newService()
	svc.Status.SetConditions(apis.Conditions{{
		Type:   apis.ConditionReady,
		Status: corev1.ConditionFalse,
	}})
	return svc
}
