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

// Package function contains utilities for Functions tests.
package function

import (
	"github.com/ghodss/yaml"
	"github.com/onsi/ginkgo"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	controllerNamespace         = "serverless-system"
	controllerConfigMapName     = "fn-config"
	controllerServiceAccountKey = "serviceAccountName"
	controllerDefaultsKey       = "defaults"
	controllerRegistryKey       = "dockerRegistry"
)

// ServiceAccountName returns the name of the ServiceAccount configured for the
// controller. This ServiceAccount is expected to exist in every namespace
// where Functions are created.
func ServiceAccountName(c kubernetes.Interface) string {
	cm := controllerConfigMap(c)
	sa, ok := cm.Data[controllerServiceAccountKey]
	if !ok {
		framework.Failf("ServiceAccount name missing from controller ConfigMap %q", cm.Name)
	}
	return sa
}

// DefaultConfig mirrors function-controller/pkg/utils.DefaultConfig to avoid
// import conflicts
type DefaultConfig struct {
	Runtime         string `json:"runtime"`
	Size            string `json:"size"`
	TimeOut         int32  `json:"timeOut"`
	FuncContentType string `json:"funcContentType"`
}

// Defaults returns the current Function defaults configured for the
// controller.
func Defaults(c kubernetes.Interface) (cfg DefaultConfig) {
	cm := controllerConfigMap(c)
	d, ok := cm.Data[controllerDefaultsKey]
	if !ok {
		framework.Failf("Defaults missing from controller ConfigMap %q", cm.Name)
	}

	if err := yaml.Unmarshal([]byte(d), &cfg); err != nil {
		framework.Failf("Unable to unmarshal defaults from controller ConfigMap: %v:", err)
	}
	return cfg
}

// SetControllerRegistry updates the URL of the container registry in the
// controller's ConfigMap and returns a function that reverts the change.
func SetControllerRegistry(c kubernetes.Interface, registryURL string) (undo func()) {
	var setRegistryURL patchConfigMapFn = func(cm *corev1.ConfigMap) revertConfigMapFn {
		originalRegistryURL := cm.Data[controllerRegistryKey]

		cm.Data[controllerRegistryKey] = registryURL

		var revert revertConfigMapFn = func(cm *corev1.ConfigMap) {
			// reset only if no other writer has modified the URL
			// since the initial patching
			if cm.Data[controllerRegistryKey] == registryURL {
				cm.Data[controllerRegistryKey] = originalRegistryURL
			}
		}

		return revert
	}

	return updateControllerConfigMap(c, setRegistryURL)
}

// patchConfigMapFn alters a ConfigMap object and return a function that
// reverts the changes on that ConfigMap.
type patchConfigMapFn func(*corev1.ConfigMap) revertConfigMapFn
type revertConfigMapFn func(*corev1.ConfigMap)

// updateControllerConfigMap updates the configuration of the Function
// controller using the given patch function and returns another function that
// can be executed to undo the changes.
func updateControllerConfigMap(c kubernetes.Interface, patch patchConfigMapFn) (undo func()) {
	cm := controllerConfigMap(c)
	revert := patch(cm)

	updateConfigMap(c, cm)

	return func() {
		ginkgo.By("reverting changes to the controller configuration")
		cm := controllerConfigMap(c)
		revert(cm)
		updateConfigMap(c, cm)
	}
}

func updateConfigMap(c kubernetes.Interface, cm *corev1.ConfigMap) {
	_, err := c.CoreV1().ConfigMaps(controllerNamespace).Update(cm)
	if err != nil {
		framework.Failf("Error updating controller ConfigMap: %v", err)
	}
}

func controllerConfigMap(c kubernetes.Interface) *corev1.ConfigMap {
	getOpts := metav1.GetOptions{}
	cm, err := c.CoreV1().ConfigMaps(controllerNamespace).Get(controllerConfigMapName, getOpts)
	if err != nil {
		framework.Failf("Error getting controller ConfigMap: %v", err)
	}

	return cm
}
