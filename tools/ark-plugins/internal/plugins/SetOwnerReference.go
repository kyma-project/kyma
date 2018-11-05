package plugins

import (
	"github.com/heptio/ark/pkg/apis/ark/v1"
	"github.com/heptio/ark/pkg/restore"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	//scapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

// SetOwnerReference is a plugin for ark to set new value of UID in metadata.ownerReference field
// based on new values of other restored objects
type SetOwnerReference struct {
	Log logrus.FieldLogger
}

// AppliesTo return list of resource kinds which should be handled by this plugin
func (p *SetOwnerReference) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{
		IncludedResources: []string{"secret"},
	}, nil
}

// Execute contains main logic for plugin
// nolint
func (p *SetOwnerReference) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
	scClient, err := scclient.NewForConfig(k8sConfig)
	sb, err := scClient.ServicecatalogV1beta1().ServiceBindings(metadata.GetNamespace()).Get(metadata.GetName(), metav1.GetOptions{})
	if err != nil {
		p.Log.Infof("Couldn't get SB %s: %v", metadata.GetName(), err)
		return item, nil, nil // there's no servicebinding with such name so the secret should not have ownerReference set
	}

	p.Log.Infof("Setting owner reference for %s %s in namespace %s", item.GetObjectKind(), metadata.GetName(), metadata.GetNamespace())

	boolTrue := true
	ownerReferences := []metav1.OwnerReference{{
		APIVersion:         "servicecatalog.k8s.io/v1beta1",
		BlockOwnerDeletion: &boolTrue,
		Controller:         &boolTrue,
		Kind:               "ServiceBinding",
		Name:               metadata.GetName(),
		UID:                sb.GetUID(),
	}}
	metadata.SetOwnerReferences(ownerReferences)

	return item, nil, nil
}
