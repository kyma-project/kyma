package httpsource

import (
	eventsourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(httpSources *eventsourcesv1alpha1.HTTPSourceList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	httpSourceUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(httpSources)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(httpSources)
	if err != nil {
		return Client{}, err
	}

	httpSourceUnstructured := &unstructured.UnstructuredList{
		Object: httpSourceUnstructuredMap,
		Items:  unstructuredItems,
	}
	httpSourceUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, httpSourceUnstructured)
	return Client{DynamicClient: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := eventsourcesv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func toUnstructuredItems(httpSourceList *eventsourcesv1alpha1.HTTPSourceList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, httpSource := range httpSourceList.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&httpSource)
		if err != nil {
			return nil, err
		}
		httpSourceUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		httpSourceUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *httpSourceUnstructured)
	}
	return unstructuredItems, nil
}
