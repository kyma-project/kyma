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

package object

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewServiceMonitor creates a ServiceMonitor object.
func NewServiceMonitor(ns, name string, opts ...ObjectOption) *monitoringv1.ServiceMonitor {
	s := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddSpecEndpoints adds the endpoints in the spec for a Servicemonitor.
func AddSpecEndpoints(endpoints ...string) ObjectOption {
	return func(o metav1.Object) {
		sm := o.(*monitoringv1.ServiceMonitor)
		eps := &sm.Spec.Endpoints
		if *eps == nil {
			*eps = sm.Spec.Endpoints
		}
		for i, ep := range endpoints {
			newEp := monitoringv1.Endpoint{
				Port: ep,
			}
			(*eps)[i] = newEp
		}
	}
}

// AddSelector adds the endpoints in the spec for a Servicemonitor.
func AddSelector(labels map[string]string) ObjectOption {
	return func(o metav1.Object) {
		sm := o.(*monitoringv1.ServiceMonitor)
		matchLabels := &sm.Spec.Selector.MatchLabels
		if *matchLabels == nil {
			*matchLabels = sm.Spec.Selector.MatchLabels
		}
		matchLabels = &labels
	}
}

// ApplyExistingServiceMonitorAttributes copies some important attributes from a given
// source ServiceMonitor to a destination ServiceMonitor.
func ApplyExistingServiceMonitorAttributes(src, dst *monitoringv1.ServiceMonitor) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion

	// Labels must be
	for _, ann := range knativeMessagingAnnotations {
		if val, ok := src.Annotations[ann]; ok {
			metav1.SetMetaDataAnnotation(&dst.ObjectMeta, ann, val)
		}
	}
}

// ApplySpec Endpoints and selector
// object.WithLabel(serviceMonitorKnativeServiceLabelKey, src.Name),
//		 = "serving.knative.dev/service"
//		serviceMonitorServiceTypeLabelKey    = "networking.internal.knative.dev/serviceType"
//		serviceMonitorServiceTypeLabelValue  = "Private"
