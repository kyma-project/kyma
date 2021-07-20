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
		It("Should sync with the Fluent Bit configuration", func() {
			By("By creating a new LoggingConfiguration")
			ctx := context.Background()

			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: ControllerNamespace,
				},
				StringData: map[string]string{
					"key": "value",
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			file := telemetryv1alpha1.FileMount{
				Name:    "myFile",
				Content: "file-content",
			}
			secretRef := telemetryv1alpha1.SecretReference{
				Name:      "my-secret",
				Namespace: ControllerNamespace,
			}
			section := telemetryv1alpha1.Section{
				Content:     FluentBitOutputConfig,
				Files:       []telemetryv1alpha1.FileMount{file},
				Environment: []telemetryv1alpha1.SecretReference{secretRef},
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

			// Fluent Bit config section should be copied to ConfigMap
			Eventually(func() string {
				cmFileName := LoggingConfigurationName + ".conf"
				configMapLookupKey := types.NamespacedName{
					Name:      FluentBitConfigMap,
					Namespace: ControllerNamespace,
				}
				var fluentBitCm corev1.ConfigMap
				err := k8sClient.Get(ctx, configMapLookupKey, &fluentBitCm)
				if err != nil {
					return err.Error()
				}
				return strings.TrimRight(fluentBitCm.Data[cmFileName], "\n")
			}, timeout, interval).Should(Equal(FluentBitOutputConfig))

			// File content should be copied to ConfigMap
			Eventually(func() string {
				filesConfigMapLookupKey := types.NamespacedName{
					Name:      FluentBitFilesConfigMap,
					Namespace: ControllerNamespace,
				}
				var filesCm corev1.ConfigMap
				err := k8sClient.Get(ctx, filesConfigMapLookupKey, &filesCm)
				if err != nil {
					return err.Error()
				}
				return filesCm.Data["myFile"]
			}, timeout, interval).Should(Equal("file-content"))

			// Secret reference should be copied to environment Secret
			Eventually(func() string {
				envSecretLookupKey := types.NamespacedName{
					Name:      FluentBitEnvSecret,
					Namespace: ControllerNamespace,
				}
				var envSecret corev1.Secret
				err := k8sClient.Get(ctx, envSecretLookupKey, &envSecret)
				if err != nil {
					return err.Error()
				}
				return string(envSecret.Data["key"])
			}, timeout, interval).Should(Equal("value"))

			// Finalizers should be added
			Eventually(func() []string {
				loggingConfigLookupKey := types.NamespacedName{
					Name:      LoggingConfigurationName,
					Namespace: ControllerNamespace,
				}
				var updatedLoggingConfiguration telemetryv1alpha1.LoggingConfiguration
				k8sClient.Get(ctx, loggingConfigLookupKey, &updatedLoggingConfiguration)
				return updatedLoggingConfiguration.Finalizers
			}, timeout, interval).Should(ContainElement(configMapFinalizer))

			Expect(k8sClient.Delete(ctx, loggingConfiguration)).Should(Succeed())
		})
	})
})
