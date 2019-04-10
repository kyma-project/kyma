package retriever

import (
	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ServiceClassByExternalID(scClient v1beta1.ServicecatalogV1beta1Interface, namespace, externalID string) (*catalog.ServiceClass, error) {
	sc, err := scClient.ServiceClasses(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "while getting service classes")
	}

	var items []catalog.ServiceClass
	for _, item := range sc.Items {
		if item.Spec.ExternalID == externalID {
			items = append(items, item)
		}
	}

	noItems := len(items)
	if noItems == 0 { // using apierrors to simplify assertion in test code
		return nil, apierrors.NewNotFound(schema.GroupResource{}, externalID)
	}

	if noItems > 1 {
		return nil, errors.Errorf("Expect one ServiceClassByExternalID with %s:%s. Found %d", catalog.FilterSpecExternalID, externalID, noItems)
	}

	return &items[0], nil
}