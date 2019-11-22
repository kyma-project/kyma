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
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	fakeservingclient "knative.dev/serving/pkg/client/injection/client/fake"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	fakesourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-sources/reconciler/testing"
)

const (
	tNs      = "testns"
	tName    = "test"
	tImg     = "sources.kyma-project.io/http:latest"
	tPort    = 8080
	tSinkURI = "http://" + tName + "-kn-channel." + tNs + ".svc.cluster.local"

	tMetricsDomain = "testing"
)

var tChLabels = labels.Set{
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
		Value: DefaultHTTPSource,
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
				NewHTTPSource(tNs, tName),
			},
			WantCreates: []runtime.Object{
				newChannelNotReady(),
				// no Service gets created until the Channel
				// becomes ready
			},
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewHTTPSource(tNs, tName,
					WithInitConditions,
					WithNoSink,
					// "Deployed" condition remains Unknown
				),
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
				NewService(tNs, tName,
					WithServiceContainer("outdated", 0, nil),
					WithServiceReady,
				),
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
				NewChannel(tNs, tName,
					WithChannelLabels(labels.Set{"not": "expected"}),
					WithChannelSinkURI(tSinkURI),
				),
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

// Deployed: True, SinkProvided: True
func newSourceDeployedWithSink() *sourcesv1alpha1.HTTPSource {
	return NewHTTPSource(tNs, tName,
		WithInitConditions,
		WithDeployed,
		WithSink(tSinkURI),
	)
}

// Deployed: True, SinkProvided: False
func newSourceDeployedWithoutSink() *sourcesv1alpha1.HTTPSource {
	return NewHTTPSource(tNs, tName,
		WithInitConditions,
		WithDeployed,
		WithNoSink,
	)
}

// Deployed: False, SinkProvided: True
func newSourceNotDeployedWithSink() *sourcesv1alpha1.HTTPSource {
	return NewHTTPSource(tNs, tName,
		WithInitConditions,
		WithNotDeployed,
		WithSink(tSinkURI),
	)
}

// addressable
func newChannelReady() *messagingv1alpha1.Channel {
	return NewChannel(tNs, tName,
		WithChannelLabels(tChLabels),
		WithChannelController(tName),
		WithChannelSinkURI(tSinkURI),
	)
}

// not addressable
func newChannelNotReady() *messagingv1alpha1.Channel {
	return NewChannel(tNs, tName,
		WithChannelLabels(tChLabels),
		WithChannelController(tName),
	)
}

// Ready: True
func newServiceReady() *servingv1alpha1.Service {
	return NewService(tNs, tName,
		WithServiceController(tName),
		WithServiceContainer(tImg, tPort, tEnvVars),
		WithServiceReady,
	)
}

// Ready: False
func newServiceNotReady() *servingv1alpha1.Service {
	return NewService(tNs, tName,
		WithServiceController(tName),
		WithServiceContainer(tImg, tPort, tEnvVars),
		WithServiceNotReady,
	)
}
