package configmap

import (
	"encoding/json"

	"github.com/pkg/errors"

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

func (c Client) List(namespace string, labelSelector string) (*corev1.ConfigMapList, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	configMapUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	return toConfigMapList(configMapUnstructured)
}

func (c Client) Get(namespace, name string) (*corev1.ConfigMap, error) {
	unstructuredConfigMap, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toConfigMap(unstructuredConfigMap)
}

func (c Client) Create(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	unstructuredObj, err := toUnstructured(configMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert cm to unstructured")
	}
	unstructuredConfigMap, err := c.DynamicClient.
		Resource(GroupVersionResource()).
		Namespace(configMap.Namespace).
		Create(unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return toConfigMap(unstructuredConfigMap)
}

func toUnstructured(configMap *corev1.ConfigMap) (*unstructured.Unstructured, error) {
	object, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toConfigMap(unstructuredConfigMap *unstructured.Unstructured) (*corev1.ConfigMap, error) {
	configMap := new(corev1.ConfigMap)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfigMap.Object, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func toConfigMapList(unstructuredList *unstructured.UnstructuredList) (*corev1.ConfigMapList, error) {
	configMapList := new(corev1.ConfigMapList)
	configMapListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(configMapListBytes, configMapList)
	if err != nil {
		return nil, err
	}
	return configMapList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  corev1.SchemeGroupVersion.Version,
		Group:    corev1.SchemeGroupVersion.Group,
		Resource: "configmaps",
	}
}
