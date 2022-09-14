//go:build unit
// +build unit

package object

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

func TestEventingBackendEqual(t *testing.T) {
	emptyBackend := eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: eventingv1alpha1.EventingBackendSpec{},
	}

	testCases := map[string]struct {
		getBackend1    func() *eventingv1alpha1.EventingBackend
		getBackend2    func() *eventingv1alpha1.EventingBackend
		expectedResult bool
	}{
		"should be unequal if labels are different": {
			getBackend1: func() *eventingv1alpha1.EventingBackend {
				b := emptyBackend.DeepCopy()
				b.Labels = map[string]string{"k1": "v1"}
				return b
			},
			getBackend2: func() *eventingv1alpha1.EventingBackend {
				return emptyBackend.DeepCopy()
			},
			expectedResult: false,
		},
		"should be equal if labels are the same": {
			getBackend1: func() *eventingv1alpha1.EventingBackend {
				b := emptyBackend.DeepCopy()
				b.Labels = map[string]string{"k1": "v1"}
				return b
			},
			getBackend2: func() *eventingv1alpha1.EventingBackend {
				b := emptyBackend.DeepCopy()
				b.Name = "bar"
				b.Labels = map[string]string{"k1": "v1"}
				return b
			},
			expectedResult: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if eventingBackendEqual(tc.getBackend1(), tc.getBackend2()) != tc.expectedResult {
				t.Errorf("expected output to be %t", tc.expectedResult)
			}
		})
	}
}

func TestPublisherProxyDeploymentEqual(t *testing.T) {
	publisherCfg := env.PublisherConfig{
		Image:          "publisher",
		PortNum:        0,
		MetricsPortNum: 0,
		ServiceAccount: "publisher-sa",
		Replicas:       1,
		RequestsCPU:    "32m",
		RequestsMemory: "64Mi",
		LimitsCPU:      "64m",
		LimitsMemory:   "128Mi",
	}
	natsConfig := env.NatsConfig{
		EventTypePrefix: "prefix",
		JSStreamName:    "kyma",
	}
	defaultNATSPublisher := deployment.NewNATSPublisherDeployment(natsConfig, publisherCfg)
	defaultBEBPublisher := deployment.NewBEBPublisherDeployment(publisherCfg)

	testCases := map[string]struct {
		getPublisher1  func() *appsv1.Deployment
		getPublisher2  func() *appsv1.Deployment
		expectedResult bool
	}{
		"should be equal if same default NATS publisher": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Name = "publisher1"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Name = "publisher2"
				return p
			},
			expectedResult: true,
		},
		"should be equal if same default BEB publisher": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultBEBPublisher.DeepCopy()
				p.Name = "publisher1"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultBEBPublisher.DeepCopy()
				p.Name = "publisher2"
				return p
			},
			expectedResult: true,
		},
		"should be unequal if publisher types are different": {
			getPublisher1: func() *appsv1.Deployment {
				return defaultBEBPublisher.DeepCopy()
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be unequal if publisher image changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Spec.Containers[0].Image = "new-publisher-img"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be unequal if env var changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Spec.Containers[0].Env[0].Value = "new-value"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be unequal if replicas changes": {
			getPublisher1: func() *appsv1.Deployment {
				replicas := int32(2)
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Replicas = &replicas
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be equal if spec annotations are nil and empty": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = nil
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = map[string]string{}
				return p
			},
			expectedResult: true,
		},
		"should be unequal if spec annotations changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = map[string]string{"key": "value1"}
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = map[string]string{"key": "value2"}
				return p
			},
			expectedResult: false,
		},
		"should be equal if spec Labels are nil and empty": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = nil
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = map[string]string{}
				return p
			},
			expectedResult: true,
		},
		"should be unequal if spec Labels changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = map[string]string{"key": "value1"}
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = map[string]string{"key": "value2"}
				return p
			},
			expectedResult: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if publisherProxyDeploymentEqual(tc.getPublisher1(), tc.getPublisher2()) != tc.expectedResult {
				t.Errorf("expected output to be %t", tc.expectedResult)
			}
		})
	}
}
