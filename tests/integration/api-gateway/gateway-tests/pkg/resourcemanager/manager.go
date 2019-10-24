package resourcemanager

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"time"
)

//CreateResource creates a given k8s resource
func CreateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, manifest unstructured.Unstructured) {
	_, err := client.Resource(resourceSchema).Namespace(namespace).Create(&manifest, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

//UpdateResource updates a given k8s resource
func UpdateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, name string, updateTo unstructured.Unstructured) {
	time.Sleep(5 * time.Second) //TODO: delete after waiting for resource creation is implemented

	toUpdate, err := client.Resource(resourceSchema).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	updateTo.SetResourceVersion(toUpdate.GetResourceVersion())
	_, err = client.Resource(resourceSchema).Namespace(namespace).Update(&updateTo, metav1.UpdateOptions{})
	if err != nil {
		panic(err)
	}
}

//DeleteResource deletes a given k8s resource
func DeleteResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := client.Resource(resourceSchema).Namespace(namespace).Delete(resourceName, deleteOptions); err != nil {
		panic(err)
	}
}
