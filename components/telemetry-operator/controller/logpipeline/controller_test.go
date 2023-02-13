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

package logpipeline

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

var _ = Describe("LogPipeline controller", Ordered, func() {
	const (
		LogPipelineName       = "log-pipeline"
		FluentBitFilterConfig = "Name   grep\nRegex   $kubernetes['labels']['app'] my-deployment"
		FluentBitOutputConfig = "Name   stdout\n"
		timeout               = time.Second * 10
		interval              = time.Millisecond * 250
	)
	var expectedFluentBitConfig = `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          log-pipeline-stdout
    emitter_storage.type  filesystem
    rule                  $log "^.*$" log-pipeline.$TAG true

[FILTER]
    name   record_modifier
    match  log-pipeline.*
    record cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    name  grep
    match log-pipeline.*
    regex $kubernetes['labels']['app'] my-deployment

[FILTER]
    name         nest
    match        log-pipeline.*
    add_prefix   __kyma__
    nested_under kubernetes
    operation    lift

[FILTER]
    name       record_modifier
    match      log-pipeline.*
    remove_key __kyma__annotations

[FILTER]
    name          nest
    match         log-pipeline.*
    nest_under    kubernetes
    operation     nest
    remove_prefix __kyma__
    wildcard      __kyma__*

[OUTPUT]
    name                     stdout
    match                    log-pipeline.*
    alias                    log-pipeline-stdout
    storage.total_limit_size 1G`

	var expectedSecret = make(map[string][]byte)
	expectedSecret["myKey"] = []byte("value")
	file := telemetryv1alpha1.FileMount{
		Name:    "myFile",
		Content: "file-content",
	}
	secretKeyRef := telemetryv1alpha1.SecretKeyRef{
		Name:      "my-secret",
		Namespace: testConfig.DaemonSet.Namespace,
		Key:       "key",
	}
	variableRefs := telemetryv1alpha1.VariableRef{
		Name:      "myKey",
		ValueFrom: telemetryv1alpha1.ValueFromSource{SecretKeyRef: &secretKeyRef},
	}
	filter := telemetryv1alpha1.Filter{
		Custom: FluentBitFilterConfig,
	}
	var logPipeline = &telemetryv1alpha1.LogPipeline{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "telemetry.kyma-project.io/v1alpha1",
			Kind:       "LogPipeline",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: LogPipelineName,
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Input: telemetryv1alpha1.Input{Application: telemetryv1alpha1.ApplicationInput{
				Namespaces: telemetryv1alpha1.InputNamespaces{
					System: true}}},
			Filters:   []telemetryv1alpha1.Filter{filter},
			Output:    telemetryv1alpha1.Output{Custom: FluentBitOutputConfig},
			Files:     []telemetryv1alpha1.FileMount{file},
			Variables: []telemetryv1alpha1.VariableRef{variableRefs},
		},
	}
	Context("On startup", Ordered, func() {
		It("Should not have any Logpipelines", func() {
			ctx := context.Background()
			var logPipelineList telemetryv1alpha1.LogPipelineList
			err := k8sClient.List(ctx, &logPipelineList)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(logPipelineList.Items)).Should(Equal(0))
		})
		It("Should not have any fluent-bit daemon set", func() {
			var fluentBitDaemonSetList appsv1.DaemonSetList
			err := k8sClient.List(ctx, &fluentBitDaemonSetList)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(fluentBitDaemonSetList.Items)).Should(Equal(0))
		})
	})
	Context("When creating a log pipeline", Ordered, func() {
		BeforeAll(func() {
			ctx := context.Background()
			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: testConfig.DaemonSet.Namespace,
				},
				StringData: map[string]string{
					"key": "value",
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
			Expect(k8sClient.Create(ctx, logPipeline)).Should(Succeed())

		})
		It("Should verify metrics from the telemetry operator are exported", func() {
			Eventually(func() bool {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return false
				}
				defer resp.Body.Close()
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.Contains(line, "telemetry_all_logpipelines") || strings.Contains(line, "telemetry_unsupported_logpipelines") {
						return true
					}
				}
				return false
			}, timeout, interval).Should(Equal(true))
		})
		It("Should have the telemetry_all_logpipelines metric set to 1", func() {
			// All log pipeline gauge should be updated
			Eventually(func() float64 {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return 0
				}
				var parser expfmt.TextParser
				mf, err := parser.TextToMetricFamilies(resp.Body)
				if err != nil {
					return 0
				}

				return *mf["telemetry_all_logpipelines"].Metric[0].Gauge.Value
			}, timeout, interval).Should(Equal(1.0))
		})
		It("Should have the telemetry_unsupported_logpipelines metric set to 1", func() {
			Eventually(func() float64 {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return 0
				}
				var parser expfmt.TextParser
				mf, err := parser.TextToMetricFamilies(resp.Body)
				if err != nil {
					return 0
				}

				return *mf["telemetry_unsupported_logpipelines"].Metric[0].Gauge.Value
			}, timeout, interval).Should(Equal(1.0))
		})
		It("Should have fluent bit config section copied to the Fluent Bit configmap", func() {
			// Fluent Bit config section should be copied to ConfigMap
			Eventually(func() string {
				cmFileName := LogPipelineName + ".conf"
				configMapLookupKey := types.NamespacedName{
					Name:      testConfig.SectionsConfigMap.Name,
					Namespace: testConfig.SectionsConfigMap.Namespace,
				}
				var fluentBitCm corev1.ConfigMap
				err := k8sClient.Get(ctx, configMapLookupKey, &fluentBitCm)
				if err != nil {
					return err.Error()
				}
				actualFluentBitConfig := strings.TrimRight(fluentBitCm.Data[cmFileName], "\n")
				return actualFluentBitConfig
			}, timeout, interval).Should(Equal(expectedFluentBitConfig))
		})
		It("Should verify files have been copied into -files configmap", func() {
			Eventually(func() string {
				filesConfigMapLookupKey := types.NamespacedName{
					Name:      testConfig.FilesConfigMap.Name,
					Namespace: testConfig.FilesConfigMap.Namespace,
				}
				var filesCm corev1.ConfigMap
				err := k8sClient.Get(ctx, filesConfigMapLookupKey, &filesCm)
				if err != nil {
					return err.Error()
				}
				return filesCm.Data["myFile"]
			}, timeout, interval).Should(Equal("file-content"))
		})

		It("Should have created flunent-bit parsers configmap", func() {
			Eventually(func() string {
				parserCmName := fmt.Sprintf("%s-parsers", testConfig.DaemonSet.Name)
				filesConfigMapLookupKey := types.NamespacedName{
					Name:      parserCmName,
					Namespace: testConfig.FilesConfigMap.Namespace,
				}
				var filesCm corev1.ConfigMap
				err := k8sClient.Get(ctx, filesConfigMapLookupKey, &filesCm)
				if err != nil {
					return err.Error()
				}
				return filesCm.Data["parsers.conf"]
			}, timeout, interval).Should(Equal(""))
		})

		It("Should verify secret reference is copied into environment secret", func() {
			Eventually(func() string {
				envSecretLookupKey := types.NamespacedName{
					Name:      testConfig.EnvSecret.Name,
					Namespace: testConfig.EnvSecret.Namespace,
				}
				var envSecret corev1.Secret
				err := k8sClient.Get(ctx, envSecretLookupKey, &envSecret)
				if err != nil {
					return err.Error()
				}
				return string(envSecret.Data["myKey"])
			}, timeout, interval).Should(Equal("value"))
		})
		It("Should have added the finalizers", func() {
			Eventually(func() []string {
				loggingConfigLookupKey := types.NamespacedName{
					Name:      LogPipelineName,
					Namespace: testConfig.DaemonSet.Namespace,
				}
				var updatedLogPipeline telemetryv1alpha1.LogPipeline
				err := k8sClient.Get(ctx, loggingConfigLookupKey, &updatedLogPipeline)
				if err != nil {
					return []string{err.Error()}
				}
				return updatedLogPipeline.Finalizers
			}, timeout, interval).Should(ContainElement("FLUENT_BIT_SECTIONS_CONFIG_MAP"))
		})
		It("Should have created a fluent-bit daemon set", func() {
			Eventually(func() int {
				var fluentBitDaemonSet appsv1.DaemonSet
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      testConfig.DaemonSet.Name,
					Namespace: testConfig.DaemonSet.Namespace,
				}, &fluentBitDaemonSet)
				if err != nil {
					return 0
				}
				return int(fluentBitDaemonSet.Generation)
			}, timeout, interval).Should(Equal(1))
		})
		It("Should have the checksum annotation set to the fluent-bit daemonset", func() {
			// Fluent Bit daemon set should have checksum annotation set
			Eventually(func() bool {
				var fluentBitDaemonSet appsv1.DaemonSet
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      testConfig.DaemonSet.Name,
					Namespace: testConfig.DaemonSet.Namespace,
				}, &fluentBitDaemonSet)
				if err != nil {
					return false
				}

				_, found := fluentBitDaemonSet.Spec.Template.Annotations["checksum/logpipeline-config"]
				return found
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting the log pipeline", Ordered, func() {

		BeforeAll(func() {
			Expect(k8sClient.Delete(ctx, logPipeline)).Should(Succeed())
		})

		It("Should have no fluent bit daemon set", func() {
			Eventually(func() int {
				var fluentBitDaemonSetList appsv1.DaemonSetList
				err := k8sClient.List(ctx, &fluentBitDaemonSetList)
				Expect(err).ShouldNot(HaveOccurred())
				return len(fluentBitDaemonSetList.Items)
			}, timeout, interval).Should(Equal(0))
		})

		It("Should reset the telemetry_all_logpipelines metric", func() {
			Eventually(func() float64 {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return 0
				}
				var parser expfmt.TextParser
				mf, err := parser.TextToMetricFamilies(resp.Body)
				if err != nil {
					return 0
				}

				return *mf["telemetry_all_logpipelines"].Metric[0].Gauge.Value
			}, timeout, interval).Should(Equal(0.0))
		})

		It("Should reset the telemetry_unsupported_logpipelines metric", func() {
			Eventually(func() float64 {
				resp, err := http.Get("http://localhost:8080/metrics")
				if err != nil {
					return 0
				}
				var parser expfmt.TextParser
				mf, err := parser.TextToMetricFamilies(resp.Body)
				if err != nil {
					return 0
				}

				return *mf["telemetry_unsupported_logpipelines"].Metric[0].Gauge.Value
			}, timeout, interval).Should(Equal(0.0))
		})
	})
})
