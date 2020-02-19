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

package plugins

import (
	"fmt"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
)

const ()

// IgnoreNatssChannelService ignores Services associated to NATSS Channels during restore.
type IgnoreByLabel struct {
	Log logrus.FieldLogger
}

func ignoredLabels() map[string]string {
	return map[string]string{
		"eventing.knative.dev/broker": "default",
	}

}

// AppliesTo returns a selector that determines what objects this plugin applies to.
func (p *IgnoreByLabel) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		ExcludedNamespaces: []string{
			"kyma-system",
			"knative-eventing",
			"knative-serving",
			"istio-system",
		},
		ExcludedResources: []string{
			"trigger.eventing.knative.dev",
		},
	}, nil
}

// Execute contains executes the plugin logic on the received object.
func (p *IgnoreByLabel) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	item := input.Item

	meta, err := apimeta.Accessor(item)
	if err != nil {
		return nil, fmt.Errorf("accessing item metadata: %s", err)
	}

	itemStr := fmt.Sprintf("%s %s/%s",
		item.GetObjectKind().GroupVersionKind().GroupKind(),
		meta.GetNamespace(),
		meta.GetName(),
	)

	restoreOutput := velero.NewRestoreItemActionExecuteOutput(input.Item)
	for key, value := range ignoredLabels() {
		if labelValue, ok := meta.GetLabels()[key]; ok && value == labelValue {
			p.Log.Infof("Ignoring restore of %s as it had label %v:%v", itemStr, key, value)
			return restoreOutput.WithoutRestore(), nil
		}
	}

	p.Log.Infof("Restoring %s", itemStr)
	return restoreOutput, nil
}
