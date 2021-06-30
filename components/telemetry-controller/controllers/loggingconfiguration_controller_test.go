/*
Copyright 2021.

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

package controllers

import (
	"context"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("LoggingConfiguration controller", func() {
	const (
		LoggingConfigurationName = "logging-configuration"
		FluentBitOutputConfig    = "[OUTPUT]\n  Name \"Null\"\n  Match *"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating LoggingConfiguration", func() {
		It("Should sync with the ConfigMap entry", func() {
			By("By creating a new LoggingConfiguration")
			ctx := context.Background()
			cmFileName := LoggingConfigurationName + ".conf"

			section := telemetryv1alpha1.Section{
				Content: FluentBitOutputConfig,
			}
			loggingConfiguration := &telemetryv1alpha1.LoggingConfiguration{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "telemetry.kyma-project.io/v1alpha1",
					Kind:       "LoggingConfiguration",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      LoggingConfigurationName,
					Namespace: ControllerNamespace,
				},
				Spec: telemetryv1alpha1.LoggingConfigurationSpec{
					Sections: []telemetryv1alpha1.Section{section},
				},
			}
			Expect(k8sClient.Create(ctx, loggingConfiguration)).Should(Succeed())

			configMapLookupKey := types.NamespacedName{
				Name:      FluentBitConfigMap,
				Namespace: ControllerNamespace,
			}
			Eventually(func() string {
				var fluentBitCm corev1.ConfigMap
				err := k8sClient.Get(ctx, configMapLookupKey, &fluentBitCm)
				if err != nil {
					return err.Error()
				}
				return strings.TrimRight(fluentBitCm.Data[cmFileName], "\n")
			}, timeout, interval).Should(Equal(FluentBitOutputConfig))

			Expect(k8sClient.Delete(ctx, loggingConfiguration)).Should(Succeed())
			//			Eventually(func() bool {
			//				var fluentBitCm corev1.ConfigMap
			//				err := k8sClient.Get(ctx, configMapLookupKey, &fluentBitCm)
			//				if err != nil {
			//					return false
			//				}
			//				if fluentBitCm.Data == nil {
			//					return false
			//				}
			//				_, hasKey := fluentBitCm.Data[cmFileName]
			//				return hasKey && fluentBitCm.Data[cmFileName] != ""
			//			}, timeout, interval).Should(BeFalse())
		})

		It("Should update finalizers", func() {
			By("By creating a new LoggingConfiguration")
			ctx := context.Background()

			section := telemetryv1alpha1.Section{
				Content: FluentBitOutputConfig,
			}
			loggingConfiguration := &telemetryv1alpha1.LoggingConfiguration{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "telemetry.kyma-project.io/v1alpha1",
					Kind:       "LoggingConfiguration",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      LoggingConfigurationName,
					Namespace: ControllerNamespace,
				},
				Spec: telemetryv1alpha1.LoggingConfigurationSpec{
					Sections: []telemetryv1alpha1.Section{section},
				},
			}
			Expect(k8sClient.Create(ctx, loggingConfiguration)).Should(Succeed())

			loggingConfigLookupKey := types.NamespacedName{
				Name:      LoggingConfigurationName,
				Namespace: ControllerNamespace,
			}

			Eventually(func() []string {
				var updatedLoggingConfiguration telemetryv1alpha1.LoggingConfiguration
				k8sClient.Get(ctx, loggingConfigLookupKey, &updatedLoggingConfiguration)
				return updatedLoggingConfiguration.Finalizers
			}, timeout, interval).Should(ContainElement(configMapFinalizer))
		})
	})
})
