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
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
)

const component = "httpsource"

// httpAdapterEnvConfig contains properties used to configure the HTTP adapter.
// These are automatically populated by envconfig.
// Calling envconfig.Process() with a prefix appends that prefix (uppercased)
// to the Go field name, e.g. HTTP_SOURCE_IMAGE.
type httpAdapterEnvConfig struct {
	// Container image
	Image string `required:"true"`
	// CloudEvents receiver port
	Port int32 `default:"8080"`
}

// Ports 9090-9091 are reserved for the Serving queue proxy (knative.dev/serving/pkg/apis/networking)
const adapterMetricsPort = 9092

// updateAdapterMetricsConfig serializes the metrics config from a ConfigMap to
// JSON and updates the existing config stored in the Reconciler.
func (r *Reconciler) updateAdapterMetricsConfig(cfg *corev1.ConfigMap) {
	metricsCfg := &metrics.ExporterOptions{
		Domain:         metrics.Domain(),
		Component:      component,
		PrometheusPort: adapterMetricsPort,
		ConfigMap:      cfg.Data,
	}

	metricsCfgJSON, err := metrics.MetricsOptionsToJson(metricsCfg)
	if err != nil {
		r.Logger.Warnw("Failed to serialize adapter metrics config to JSON",
			"configmap", cfg.Name, "error", err)
		return
	}

	r.adapterMetricsCfg = metricsCfgJSON

	r.Logger.Infow("Updated adapter metrics config from ConfigMap", "configmap", cfg.Name)
}

// updateAdapterLoggingConfig serializes the logging config from a ConfigMap to
// JSON and updates the existing config stored in the Reconciler.
func (r *Reconciler) updateAdapterLoggingConfig(cfg *corev1.ConfigMap) {
	logCfg, err := logging.NewConfigFromConfigMap(cfg)
	if err != nil {
		r.Logger.Warnw("Failed to create adapter logging config from ConfigMap",
			"configmap", cfg.Name, "error", err)
		return
	}

	logCfgJSON, err := logging.LoggingConfigToJson(logCfg)
	if err != nil {
		r.Logger.Warnw("Failed to serialize adapter logging config to JSON",
			"configmap", cfg.Name, "error", err)
		return
	}

	r.adapterLoggingCfg = logCfgJSON

	r.Logger.Infow("Updated adapter logging config from ConfigMap", "configmap", cfg.Name)
}
