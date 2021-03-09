package process

import (
	"testing"

	knativeapis "knative.dev/pkg/apis"

	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/processtest"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestConvertTriggersToKymaSubs(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	e2eSetup := newE2ESetup()

	p := &Process{
		Steps:           nil,
		ReleaseName:     "release",
		BEBNamespace:    "ns",
		EventingBackend: "nats",
		EventTypePrefix: "prefix",
		Clients:         Clients{},
		Logger:          logrus.New(),
		State:           State{},
	}
	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Convert triggers to Kyma subscriptions", func(t *testing.T) {
		saveCurrentState := NewSaveCurrentState(p)
		convertTriggersToKymaSubs := NewConvertTriggersToKymaSubscriptions(p)
		p.Steps = []Step{
			saveCurrentState,
			convertTriggersToKymaSubs,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())
		// Check for Kyma subscriptions
		for _, trigger := range e2eSetup.triggers.Items {
			gotKymaSub, err := p.Clients.KymaSubscription.Get(trigger.Namespace, trigger.Name)
			g.Expect(err).Should(gomega.BeNil())
			g.Expect(gotKymaSub.Name).To(gomega.Equal(trigger.Name))
			if trigger.Spec.Filter != nil {
				attributes := *trigger.Spec.Filter.Attributes
				expectedSub := func() *kymaeventingv1alpha1.Subscription {
					sub := processtest.NewKymaSubscription(trigger.Name, trigger.Namespace, processtest.WithSink, processtest.WithDefaultProtocolSetting)
					processtest.WithFilters(attributes[eventTypeKey], "ns", attributes[eventVersionKey], attributes[eventSourceKey], &sub)
					return &sub
				}()
				g.Expect(gotKymaSub.Spec).To(gomega.Equal(expectedSub.Spec))
			}
		}
	})
}

func TestNewSubscription(t *testing.T) {
	testCases := []struct {
		name            string
		inputTrigger    eventingv1alpha1.Trigger
		eventingBackend string
		application     applicationconnectorv1alpha1.Application
		expectedSub     kymaeventingv1alpha1.Subscription
	}{
		{
			name:            "application in trigger source exists",
			inputTrigger:    processtest.NewTrigger("trigger1", "ns", "footype", "app1", "v1"),
			eventingBackend: "nats",
			application:     processtest.NewApp("app1"),
			expectedSub: func() kymaeventingv1alpha1.Subscription {
				sub := processtest.NewKymaSubscription("trigger1", "ns", processtest.WithSink, processtest.WithDefaultProtocolSetting)
				processtest.WithFilters("footype", "foons", "v1", "app1", &sub)
				return sub
			}(),
		},
		{
			name:            "application in trigger source doesn't exist",
			inputTrigger:    processtest.NewTrigger("trigger1", "ns", "footype", "app1", "v1"),
			eventingBackend: "nats",
			application:     processtest.NewApp("does-exist"),
			expectedSub: func() kymaeventingv1alpha1.Subscription {
				sub := processtest.NewKymaSubscription("trigger1", "ns", processtest.WithSink, processtest.WithDefaultProtocolSetting)
				processtest.WithFilters("footype", "foons", "v1", "app1", &sub)
				return sub
			}(),
		},
		{
			name:            "trigger filters are empty",
			inputTrigger:    processtest.NewTriggerWithoutFilter("trigger1", "ns"),
			eventingBackend: "nats",
			expectedSub:     processtest.NewKymaSubscription("trigger1", "ns", processtest.WithDefaultProtocolSetting),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			convertStep := ConvertTriggersToKymaSubscriptions{
				name: "convert triggers to kyma subs",
				process: &Process{
					Steps:           nil,
					ReleaseName:     "release",
					BEBNamespace:    "foons",
					EventingBackend: "",
					EventTypePrefix: "prefix",
					Clients:         Clients{},
					Logger:          nil,
					State: State{
						Triggers: nil,
						Applications: &applicationconnectorv1alpha1.ApplicationList{
							Items: []applicationconnectorv1alpha1.Application{
								tc.application,
							},
						},
					},
				},
			}
			g := gomega.NewGomegaWithT(t)
			gotKymaSub := convertStep.NewSubscription(&tc.inputTrigger)
			g.Expect(gotKymaSub.ObjectMeta).To(gomega.Equal(tc.inputTrigger.ObjectMeta))
			g.Expect(gotKymaSub.Spec).To(gomega.Equal(tc.expectedSub.Spec))
		})
	}
}

func TestRewriteSink(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	testCases := []struct {
		name           string
		inputSink      *knativeapis.URL
		inputNamespace string
		expectedSink   string
	}{
		{
			name: "with protocol and cluster-local sink",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("http://test.ns.svc.cluster.local:8080/endpoint")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "ns",
			expectedSink:   "http://test.ns.svc.cluster.local:8080/endpoint",
		},
		{
			name: "with protocol and without cluster-local sink",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("http://test.ns")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "ns",
			expectedSink:   "http://test.ns.svc.cluster.local",
		},
		{
			name: "with protocol, without cluster-local sink and with ns mismatch",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("http://test.ns")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "different",
			expectedSink:   "http://test.ns",
		},
		{
			name: "with protocol, without cluster-local sink and port 8080",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("http://test.ns:8080/endpoint")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "ns",
			expectedSink:   "http://test.ns.svc.cluster.local:8080/endpoint",
		},
		{
			name: "with protocol, port=8080 and without cluster-local sink but with ns mismatch",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("http://test.ns:8080/endpoint")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "different",
			expectedSink:   "http://test.ns:8080/endpoint",
		},
		{
			name: "without protocol and cluster-local sink",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("test.ns")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "ns",
			expectedSink:   "test.ns",
		},
		{
			name: "with something external",
			inputSink: func() *knativeapis.URL {
				url, err := knativeapis.ParseURL("http://function.com/execute")
				g.Expect(err).Should(gomega.BeNil())
				return url
			}(),
			inputNamespace: "ns",
			expectedSink:   "http://function.com/execute",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotSink := rewriteSink(tc.inputSink, tc.inputNamespace)
			g.Expect(gotSink).To(gomega.Equal(tc.expectedSink))
		})
	}
}
