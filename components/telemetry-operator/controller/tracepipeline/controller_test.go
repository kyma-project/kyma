package tracepipeline

import (
	"context"
	"time"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Deploying a TracePipeline", func() {
	const (
		timeout  = time.Second * 100
		interval = time.Millisecond * 250
	)

	When("creating TracePipeline", func() {
		ctx := context.Background()
		kymaSystemNamespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kyma-system",
			},
		}
		otelCollectorDeploymentLookupKey := types.NamespacedName{
			Name:      "opentelemetry-collector",
			Namespace: "kyma-system",
		}
		otelCollectorConfigMapLookupKey := types.NamespacedName{
			Name:      "opentelemetry-collector-config",
			Namespace: "kyma-system",
		}
		tracePipeline := &telemetryv1alpha1.TracePipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-logpipeline",
			},
			Spec: telemetryv1alpha1.TracePipelineSpec{
				Output: telemetryv1alpha1.TracePipelineOutput{
					Otlp: telemetryv1alpha1.OtlpOutput{
						Endpoint: telemetryv1alpha1.ValueType{Value: "localhost"},
					},
				},
			},
		}

		It("creates OpenTelemetry Collector resources", func() {
			Expect(k8sClient.Create(ctx, kymaSystemNamespace)).Should(Succeed())
			Expect(k8sClient.Create(ctx, tracePipeline)).Should(Succeed())

			Eventually(func() error {
				var otelCollectorDeployment appsv1.Deployment
				return k8sClient.Get(ctx, otelCollectorDeploymentLookupKey, &otelCollectorDeployment)
			}, timeout, interval).Should(BeNil())

			Eventually(func() error {
				var otelCollectorService v1.Service
				return k8sClient.Get(ctx, otelCollectorDeploymentLookupKey, &otelCollectorService)
			}, timeout, interval).Should(BeNil())

			Eventually(func() error {
				var otelCollectorConfigMap v1.ConfigMap
				return k8sClient.Get(ctx, otelCollectorConfigMapLookupKey, &otelCollectorConfigMap)
			}, timeout, interval).Should(BeNil())

			Expect(k8sClient.Delete(ctx, tracePipeline)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, kymaSystemNamespace)).Should(Succeed())
		})
	})
})
