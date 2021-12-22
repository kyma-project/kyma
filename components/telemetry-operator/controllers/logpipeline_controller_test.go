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
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("LogPipeline controller", func() {
	const (
		LogPipelineName       = "log-pipeline"
		FluentBitParserConfig = "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
		FluentBitFilterConifg = "Name   grep\nMatch   *\nRegex   $kubernetes['labels']['app'] my-deployment"
		FluentBitOutputConfig = "Name   stdout\nMatch   *"
		timeout               = time.Second * 10
		interval              = time.Millisecond * 250
	)
	var expectedFluentBitConfig = fmt.Sprintf("[PARSER]\n%s\n\n[FILTER]\n%s\n\n[OUTPUT]\n%s", FluentBitParserConfig, FluentBitFilterConifg, FluentBitOutputConfig)

	Context("When updating LogPipeline", func() {
		It("Should sync with the Fluent Bit configuration", func() {
			By("By creating a new LogPipeline")
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

			podLabels := map[string]string{
				"app.kubernetes.io/instance": "logging",
				"app.kubernetes.io/name":     "fluent-bit",
			}
			container := corev1.Container{
				Name:  "fluent-bit",
				Image: "fluent-bit",
			}
			fluentBitDs := &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "DaemonSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      FluentBitDaemonSet,
					Namespace: ControllerNamespace,
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: podLabels,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: podLabels,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								container,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, fluentBitDs)).Should(Succeed())

			fluentBitPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      FluentBitDaemonSet + "-123",
					Namespace: ControllerNamespace,
					Labels:    podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						container,
					},
				},
			}
			Expect(k8sClient.Create(ctx, fluentBitPod)).Should(Succeed())

			file := telemetryv1alpha1.FileMount{
				Name:    "myFile",
				Content: "file-content",
			}
			secretRef := telemetryv1alpha1.SecretReference{
				Name:      "my-secret",
				Namespace: ControllerNamespace,
			}
			parser := telemetryv1alpha1.Parser{
				Content: FluentBitParserConfig,
			}
			filter := telemetryv1alpha1.Filter{
				Content: FluentBitFilterConifg,
			}
			output := telemetryv1alpha1.Output{
				Content: FluentBitOutputConfig,
			}
			loggingConfiguration := &telemetryv1alpha1.LogPipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "telemetry.kyma-project.io/v1alpha1",
					Kind:       "LogPipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: LogPipelineName,
				},
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Parsers:    []telemetryv1alpha1.Parser{parser},
					Filters:    []telemetryv1alpha1.Filter{filter},
					Outputs:    []telemetryv1alpha1.Output{output},
					Files:      []telemetryv1alpha1.FileMount{file},
					SecretRefs: []telemetryv1alpha1.SecretReference{secretRef},
				},
			}
			Expect(k8sClient.Create(ctx, loggingConfiguration)).Should(Succeed())

			// Fluent Bit config section should be copied to ConfigMap
			Eventually(func() string {
				cmFileName := LogPipelineName + ".conf"
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
			}, timeout, interval).Should(Equal(expectedFluentBitConfig))

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
					Name:      LogPipelineName,
					Namespace: ControllerNamespace,
				}
				var updatedLogPipeline telemetryv1alpha1.LogPipeline
				err := k8sClient.Get(ctx, loggingConfigLookupKey, &updatedLogPipeline)
				if err != nil {
					return []string{err.Error()}
				}
				return updatedLogPipeline.Finalizers
			}, timeout, interval).Should(ContainElement(configMapFinalizer))

			Expect(k8sClient.Delete(ctx, loggingConfiguration)).Should(Succeed())

			// Fluent Bit pods should have been deleted for restart
			Eventually(func() int {
				var fluentBitPods corev1.PodList
				err := k8sClient.List(ctx, &fluentBitPods, client.InNamespace(ControllerNamespace), client.MatchingLabels(podLabels))
				if err != nil {
					return 1
				}
				return len(fluentBitPods.Items)
			}, timeout, interval).Should(Equal(0))
		})
	})
})
