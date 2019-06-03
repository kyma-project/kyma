package plugins

import (
	"github.com/pkg/errors"

	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	scApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

// SetOwnerReference is a plugin for velero to set new value of UID in metadata.ownerReference field
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

	scClient, err := p.inClusterServiceCatalogClient()
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating ServiceCatalog client")
	}

	// Secret's name and service binding's name are not always equal (they are when created in kyma's console, but can be different when created by from yaml's)
	// Searching service binding by name is a workaround before https://github.com/heptio/velero/issues/965 will be resolved
	sb, err := scClient.ServiceBindings(metadata.GetNamespace()).Get(metadata.GetName(), metav1.GetOptions{})
	switch {
	case err == nil:
	case apierrors.IsNotFound(err):
		// there's no servicebinding with such name so the secret should not have ownerReference set
		p.Log.Infof("Couldn't get SB %s: %v", metadata.GetName(), err)
		return item, nil, nil
	default:
		return item, err, nil
	}

	p.Log.Infof("Setting owner reference for %s %s in namespace %s", item.GetObjectKind(), metadata.GetName(), metadata.GetNamespace())

	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(sb, bindingControllerKind),
	}
	metadata.SetOwnerReferences(ownerReferences)

	return item, nil, nil
}

// bindingControllerKind contains the schema.GroupVersionKind for ServiceCatalog controller type.
// see: https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.34/pkg/controller/controller_binding.go#L66-L67
var bindingControllerKind = scApi.SchemeGroupVersion.WithKind("ServiceBinding")

func (p *SetOwnerReference) inClusterServiceCatalogClient() (v1beta1.ServicecatalogV1beta1Interface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	scClient, err := v1beta1.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return scClient, nil
}
