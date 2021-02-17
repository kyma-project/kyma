package namespace

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	DynamicClient dynamic.Interface
}

func NewClient(dInf dynamic.Interface) Client {
	return Client{DynamicClient: dInf}
}

func (c Client) List() (*corev1.NamespaceList, error) {
	namespacesUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toNamespaceList(namespacesUnstructured)
}

func (c Client) Get(name string) (*corev1.Namespace, error) {
	unstructuredNamespace, err := c.DynamicClient.Resource(GroupVersionResource()).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toNamespace(unstructuredNamespace)
}

func (c Client) Patch(name string, data []byte) (*corev1.Namespace, error) {
	unstructuredNamespace, err := c.DynamicClient.Resource(GroupVersionResource()).Patch(name, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return toNamespace(unstructuredNamespace)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toNamespace(unstructuredDeployment *unstructured.Unstructured) (*corev1.Namespace, error) {
	namespace := new(corev1.Namespace)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func toNamespaceList(unstructuredList *unstructured.UnstructuredList) (*corev1.NamespaceList, error) {
	namespaceList := new(corev1.NamespaceList)
	namespaceListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(namespaceListBytes, namespaceList)
	if err != nil {
		return nil, err
	}
	return namespaceList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  corev1.SchemeGroupVersion.Version,
		Group:    corev1.SchemeGroupVersion.Group,
		Resource: "namespaces",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: corev1.SchemeGroupVersion.Version,
		Group:   corev1.SchemeGroupVersion.Group,
		Kind:    "Namespace",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: corev1.SchemeGroupVersion.Version,
		Group:   corev1.SchemeGroupVersion.Group,
		Kind:    "NamespaceList",
	}
}
