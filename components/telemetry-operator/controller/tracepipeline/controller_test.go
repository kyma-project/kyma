package tracepipeline

import (
	"context"
	"fmt"
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
		otelCollectorResourceLookupKey := types.NamespacedName{
			Name:      "telemetry-trace-collector",
			Namespace: "kyma-system",
		}
		tracePipeline := &telemetryv1alpha1.TracePipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-tracepipeline",
			},
			Spec: telemetryv1alpha1.TracePipelineSpec{
				Output: telemetryv1alpha1.TracePipelineOutput{
					Otlp: &telemetryv1alpha1.OtlpOutput{
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
				if err := k8sClient.Get(ctx, otelCollectorResourceLookupKey, &otelCollectorDeployment); err != nil {
					return err
				}
				if err := validateOwnerReferences(otelCollectorDeployment.OwnerReferences); err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(BeNil())

			Eventually(func() error {
				var otelCollectorService v1.Service
				if err := k8sClient.Get(ctx, otelCollectorResourceLookupKey, &otelCollectorService); err != nil {
					return err
				}
				if err := validateOwnerReferences(otelCollectorService.OwnerReferences); err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(BeNil())

			Eventually(func() error {
				var otelCollectorConfigMap v1.ConfigMap
				if err := k8sClient.Get(ctx, otelCollectorResourceLookupKey, &otelCollectorConfigMap); err != nil {
					return err
				}
				if err := validateOwnerReferences(otelCollectorConfigMap.OwnerReferences); err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(BeNil())

			Eventually(func() error {
				var otelCollectorSecret v1.Secret
				if err := k8sClient.Get(ctx, otelCollectorResourceLookupKey, &otelCollectorSecret); err != nil {
					return err
				}
				if err := validateOwnerReferences(otelCollectorSecret.OwnerReferences); err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(BeNil())

			Expect(k8sClient.Delete(ctx, tracePipeline)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, kymaSystemNamespace)).Should(Succeed())
		})
	})
})

func validateOwnerReferences(ownerRefernces []metav1.OwnerReference) error {
	if len(ownerRefernces) != 1 {
		return fmt.Errorf("unexpected number of owner references: %d", len(ownerRefernces))
	}
	ownerReference := ownerRefernces[0]

	if ownerReference.Kind != "TracePipeline" {
		return fmt.Errorf("unexpected owner reference type: %s", ownerReference.Kind)
	}

	if ownerReference.Name != "test-tracepipeline" {
		return fmt.Errorf("unexpected owner reference name: %s", ownerReference.Kind)
	}

	if !*ownerReference.BlockOwnerDeletion {
		return fmt.Errorf("owner reference does not block owner deletion")
	}
	return nil
}
