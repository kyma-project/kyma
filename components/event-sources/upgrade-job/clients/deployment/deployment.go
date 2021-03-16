package deployment

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime/schema"

	appsv1 "k8s.io/api/apps/v1"
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

func (c Client) List(namespace string, labelSelector string) (*appsv1.DeploymentList, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	deploymentsUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	return toDeploymentList(deploymentsUnstructured)
}

func (c Client) Get(namespace, name string) (*appsv1.Deployment, error) {
	unstructuredDeployment, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toDeployment(unstructuredDeployment)
}

func (c Client) Patch(namespace, name string, data []byte) (*appsv1.Deployment, error) {
	unstructuredDeployment, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Patch(name, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return toDeployment(unstructuredDeployment)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toDeployment(unstructuredDeployment *unstructured.Unstructured) (*appsv1.Deployment, error) {
	deployment := new(appsv1.Deployment)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func toDeploymentList(unstructuredList *unstructured.UnstructuredList) (*appsv1.DeploymentList, error) {
	deploymentList := new(appsv1.DeploymentList)
	deploymentListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(deploymentListBytes, deploymentList)
	if err != nil {
		return nil, err
	}
	return deploymentList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  appsv1.SchemeGroupVersion.Version,
		Group:    appsv1.SchemeGroupVersion.Group,
		Resource: "deployments",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: appsv1.SchemeGroupVersion.Version,
		Group:   appsv1.SchemeGroupVersion.Group,
		Kind:    "Deployment",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: appsv1.SchemeGroupVersion.Version,
		Group:   appsv1.SchemeGroupVersion.Group,
		Kind:    "DeploymentList",
	}
}
