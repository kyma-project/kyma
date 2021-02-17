package httpsource

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	eventsourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
)

type Client struct {
	DynamicClient dynamic.Interface
}

func NewClient(dInf dynamic.Interface) Client {
	return Client{DynamicClient: dInf}
}

func (c Client) List() (*eventsourcesv1alpha1.HTTPSourceList, error) {
	httpSourcesUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toHTTPSourceList(httpSourcesUnstructured)
}

func (c Client) Get(namespace, name string) (*eventsourcesv1alpha1.HTTPSource, error) {
	unstructuredHTTPSource, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toHTTPSource(unstructuredHTTPSource)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toHTTPSource(unstructuredDeployment *unstructured.Unstructured) (*eventsourcesv1alpha1.HTTPSource, error) {
	httpSource := new(eventsourcesv1alpha1.HTTPSource)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, httpSource)
	if err != nil {
		return nil, err
	}
	return httpSource, nil
}

func toHTTPSourceList(unstructuredList *unstructured.UnstructuredList) (*eventsourcesv1alpha1.HTTPSourceList, error) {
	httpSourceList := new(eventsourcesv1alpha1.HTTPSourceList)
	httpSourceListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(httpSourceListBytes, httpSourceList)
	if err != nil {
		return nil, err
	}
	return httpSourceList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventsourcesv1alpha1.SchemeGroupVersion.Version,
		Group:    eventsourcesv1alpha1.SchemeGroupVersion.Group,
		Resource: "httpsources",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: eventsourcesv1alpha1.SchemeGroupVersion.Version,
		Group:   eventsourcesv1alpha1.SchemeGroupVersion.Group,
		Kind:    "HTTPSource",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: eventsourcesv1alpha1.SchemeGroupVersion.Version,
		Group:   eventsourcesv1alpha1.SchemeGroupVersion.Group,
		Kind:    "HTTPSourceList",
	}
}
