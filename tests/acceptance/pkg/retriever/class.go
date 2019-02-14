package retriever

import (
	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ServiceClassByExternalID(scClient v1beta1.ServicecatalogV1beta1Interface, namespace, externalID string) (*catalog.ServiceClass, error) {
	fieldSet := fields.Set{
		catalog.FilterSpecExternalID: externalID,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := metav1.ListOptions{FieldSelector: fieldSelector}

	sc, err := scClient.ServiceClasses(namespace).List(listOpts)
	if err != nil {
		return nil, errors.Wrap(err, "while getting service class")
	}

	noItems := len(sc.Items)
	if noItems == 0 { // using apierrors to simplify assertion in test code
		return nil, apierrors.NewNotFound(schema.GroupResource{}, externalID)
	}

	if noItems > 1 {
		return nil, errors.Errorf("Expect one ServiceClassByExternalID with %s:%s. Found %d", catalog.FilterSpecExternalID, externalID, noItems)
	}

	return &sc.Items[0], nil
}
