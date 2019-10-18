package resourcemanager

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"time"
)

//CreateResource creates a given k8s resource
func CreateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, manifest unstructured.Unstructured) {
	fmt.Println("Creating resource...")
	result, err := client.Resource(resourceSchema).Namespace(namespace).Create(&manifest, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created resource %q.\n", result.GetName())
}

//UpdateResource updates a given k8s resource
func UpdateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, name string, updateTo unstructured.Unstructured) {
	fmt.Println("Updating resource...")
	time.Sleep(5*time.Second)

	toUpdate, err:=client.Resource(resourceSchema).Namespace(namespace).Get(name,metav1.GetOptions{})
	updateTo.SetResourceVersion(toUpdate.GetResourceVersion())
	fmt.Printf("Update to: %q\n",updateTo)

	result, err:= client.Resource(resourceSchema).Namespace(namespace).Update(&updateTo,metav1.UpdateOptions{})
	if err!=nil {
		panic(err)
	}
	fmt.Printf("Updated resource %q.\n", result.GetName())
}

//DeleteResource deletes a given k8s resource
func DeleteResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) {
	fmt.Println("Deleting resource...")
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := client.Resource(resourceSchema).Namespace(namespace).Delete(resourceName, deleteOptions); err != nil {
		panic(err)
	}

	fmt.Printf("Deleted resource %q.\n", resourceName)
}
