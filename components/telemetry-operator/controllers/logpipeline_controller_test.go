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
	"bufio"
	"context"
	"net/http"
	"strings"
	"time"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("LogPipeline controller", func() {
	const (
		LogPipelineName                = "log-pipeline"
		FluentBitParserConfig          = "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
		FluentBitMultiLineParserConfig = "Name          multiline-custom-regex\nType          regex\nFlush_timeout 1000\nRule      \"start_state\"   \"/(Dec \\d+ \\d+\\:\\d+\\:\\d+)(.*)/\"  \"cont\"\nRule      \"cont\"          \"/^\\s+at.*/\"                     \"cont\""
		FluentBitFilterConfig          = "Name   grep\nRegex   $kubernetes['labels']['app'] my-deployment"
		FluentBitOutputConfig          = "Name   stdout\n"
		timeout                        = time.Second * 10
		interval                       = time.Millisecond * 250
	)
	var expectedFluentBitConfig = `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    Rule                  $log "^.*$" log-pipeline.$TAG true
    Emitter_Name          log-pipeline
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M

[FILTER]
    name                  record_modifier
    match                 log-pipeline.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    match log-pipeline.*
    name grep
    regex $kubernetes['labels']['app'] my-deployment

[OUTPUT]
    match log-pipeline.*
    name stdout
    storage.total_limit_size 1G`

	var expectedFluentBitParserConfig = `[PARSER]
    Name   dummy_test
    Format   regex
    Regex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$

[MULTILINE_PARSER]
    Name          multiline-custom-regex
    Type          regex
    Flush_timeout 1000
    Rule      "start_state"   "/(Dec \d+ \d+\:\d+\:\d+)(.*)/"  "cont"
    Rule      "cont"          "/^\s+at.*/"                     "cont"`

	var expectedSecret = make(map[string][]byte)
	expectedSecret["myKey"] = []byte("value")
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
					Namespace: daemonSetConfig.FluentBitDaemonSetName.Namespace,
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
					Name:      daemonSetConfig.FluentBitDaemonSetName.Name,
					Namespace: daemonSetConfig.FluentBitDaemonSetName.Namespace,
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
					Name:      daemonSetConfig.FluentBitDaemonSetName.Name + "-123",
					Namespace: daemonSetConfig.FluentBitDaemonSetName.Namespace,
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
			secretRef := telemetryv1alpha1.SecretKeyRef{
				Name:      "my-secret",
				Namespace: daemonSetConfig.FluentBitDaemonSetName.Namespace,
				Key:       "key",
			}
			variableRefs := telemetryv1alpha1.VariableReference{
				Name:      "myKey",
				ValueFrom: telemetryv1alpha1.ValueFromType{SecretKey: secretRef},
			}
			parser := telemetryv1alpha1.Parser{
				Content: FluentBitParserConfig,
			}
			multiLineParser := telemetryv1alpha1.MultiLineParser{
				Content: FluentBitMultiLineParserConfig,
			}
			filter := telemetryv1alpha1.Filter{
				Custom: FluentBitFilterConfig,
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
					Parsers:          []telemetryv1alpha1.Parser{parser},
					MultiLineParsers: []telemetryv1alpha1.MultiLineParser{multiLineParser},
					Filters:          []telemetryv1alpha1.Filter{filter},
					Output:           telemetryv1alpha1.Output{Custom: FluentBitOutputConfig},
					Files:            []telemetryv1alpha1.FileMount{file},
					Variables:        []telemetryv1alpha1.VariableReference{variableRefs},
				},
			}
			Expect(k8sClient.Create(ctx, loggingConfiguration)).Should(Succeed())

			// Fluent Bit config section should be copied to ConfigMap
			Eventually(func() string {
				cmFileName := LogPipelineName + ".conf"
				configMapLookupKey := types.NamespacedName{
					Name:      daemonSetConfig.FluentBitSectionsConfigMap.Name,
					Namespace: daemonSetConfig.FluentBitSectionsConfigMap.Namespace,
				}
				var fluentBitCm corev1.ConfigMap
				err := k8sClient.Get(ctx, configMapLookupKey, &fluentBitCm)
				if err != nil {
					return err.Error()
				}
				actualFluentBitConfig := strings.TrimRight(fluentBitCm.Data[cmFileName], "\n")
				return actualFluentBitConfig
			}, timeout, interval).Should(Equal(expectedFluentBitConfig))

			// Fluent Bit parsers config should be copied to ConfigMap
			Eventually(func() string {
				cmFileName := "parsers.conf"
				configMapLookupKey := types.NamespacedName{
					Name:      daemonSetConfig.FluentBitParsersConfigMap.Name,
					Namespace: daemonSetConfig.FluentBitParsersConfigMap.Namespace,
				}
				var fluentBitParsersCm corev1.ConfigMap
				err := k8sClient.Get(ctx, configMapLookupKey, &fluentBitParsersCm)
				if err != nil {
					return err.Error()
				}
				return strings.TrimRight(fluentBitParsersCm.Data[cmFileName], "\n")
			}, timeout, interval).Should(Equal(expectedFluentBitParserConfig))

			// File content should be copied to ConfigMap
			Eventually(func() string {
				filesConfigMapLookupKey := types.NamespacedName{
					Name:      daemonSetConfig.FluentBitFilesConfigMap.Name,
					Namespace: daemonSetConfig.FluentBitFilesConfigMap.Namespace,
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
					Name:      daemonSetConfig.FluentBitEnvSecret.Name,
					Namespace: daemonSetConfig.FluentBitEnvSecret.Namespace,
				}
				var envSecret corev1.Secret
				err := k8sClient.Get(ctx, envSecretLookupKey, &envSecret)
				if err != nil {
					return err.Error()
				}
				return string(envSecret.Data["myKey"])
			}, timeout, interval).Should(Equal("value"))

			// Finalizers should be added
			Eventually(func() []string {
				loggingConfigLookupKey := types.NamespacedName{
					Name:      LogPipelineName,
					Namespace: daemonSetConfig.FluentBitDaemonSetName.Namespace,
				}
				var updatedLogPipeline telemetryv1alpha1.LogPipeline
				err := k8sClient.Get(ctx, loggingConfigLookupKey, &updatedLogPipeline)
				if err != nil {
					return []string{err.Error()}
				}
				return updatedLogPipeline.Finalizers
			}, timeout, interval).Should(ContainElement("FLUENT_BIT_SECTIONS_CONFIG_MAP"))

			Expect(k8sClient.Delete(ctx, loggingConfiguration)).Should(Succeed())

			// Fluent Bit daemon set should rollout-restarted (generation changes from 1 to 3)
			Eventually(func() int {
				var fluentBitDaemonSet appsv1.DaemonSet
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      daemonSetConfig.FluentBitDaemonSetName.Name,
					Namespace: daemonSetConfig.FluentBitDaemonSetName.Namespace,
				}, &fluentBitDaemonSet)
				if err != nil {
					return 0
				}
				return int(fluentBitDaemonSet.Generation)
			}, timeout, interval).Should(Equal(2))

			// Custom metrics should be exported
			Eventually(func() bool {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return false
				}
				defer resp.Body.Close()
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.Contains(line, "telemetry_fluentbit_restarts_total") {
						return true
					}
				}
				return false
			}, timeout, interval).Should(Equal(true))

			Eventually(func() bool {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return false
				}
				defer resp.Body.Close()
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.Contains(line, "telemetry_plugins_unsupported_total") {
						return true
					}
				}
				return false
			}, timeout, interval).Should(Equal(true))
		})
	})
})
