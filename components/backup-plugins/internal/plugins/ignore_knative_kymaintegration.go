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
	"strings"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
)

const (
	integrationNamespace = "kyma-integration"
	ignoredLabelPrefix   = "serving.knative.dev/"
)

// IgnoreKnative ignores Services associated with knative services
type IgnoreKnative struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a selector that determines what objects this plugin applies to.
func (p *IgnoreKnative) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedNamespaces: []string{integrationNamespace},
	}, nil
}

// Execute contains executes the plugin logic on the received object.
func (p *IgnoreKnative) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
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

	for label, _ := range meta.GetLabels() {
		if strings.HasPrefix(label, ignoredLabelPrefix) {
			p.Log.Infof("Ignoring restore of %s", itemStr)
			return restoreOutput.WithoutRestore(), nil
		}
	}

	p.Log.Infof("Restoring %s", itemStr)
	return restoreOutput, nil
}
