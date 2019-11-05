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
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
)

const natssChannelLabel = "messaging.knative.dev/role=natss-channel"

// IgnoreNatssChannelService ignores Services associated to NATSS Channels during restore.
type IgnoreNatssChannelService struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a selector that determines what objects this plugin applies to.
func (p *IgnoreNatssChannelService) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"services"},
		//LabelSelector:     natssChannelLabel,
	}, nil
}

// Execute contains executes the plugin logic on the received object.
func (p *IgnoreNatssChannelService) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.Log.Infof("Are we ignoring NatssChannel!!")
	item := input.Item
	meta, err := apimeta.Accessor(item)
	if err != nil {
		p.Log.Errorf("Error in Accessor: %v", err)
		return nil, err
	}
	labels := meta.GetLabels()
	if val, ok := labels["messaging.knative.dev/role"]; ok && val == "natss-channel" {
		p.Log.Infof("Ignoring restore of %s %s/%s", item.GetObjectKind(), meta.GetNamespace(), meta.GetName())
		return velero.NewRestoreItemActionExecuteOutput(input.Item).WithoutRestore(), nil
	}
	p.Log.Infof("Restoring %s %s/%s", item.GetObjectKind(), meta.GetNamespace(), meta.GetName())
	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
}
