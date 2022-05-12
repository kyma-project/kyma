package resource

import (
	"log"

	"github.com/avast/retry-go"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"time"
)

//Manager .
type Manager struct {
	RetryOptions []retry.Option
}

//CreateResource creates a given k8s resource
func (m *Manager) CreateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, manifest unstructured.Unstructured) {
	panicOnErr(retry.Do(func() error {
		if _, err := client.Resource(resourceSchema).Namespace(namespace).Create(&manifest, metav1.CreateOptions{}); err != nil {
			log.Printf("Error: %+v", err)
			return err
		}
		return nil
	}, m.RetryOptions...))
}

//UpdateResource updates a given k8s resource
func (m *Manager) UpdateResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, name string, updateTo unstructured.Unstructured) {
	panicOnErr(retry.Do(func() error {
		time.Sleep(5 * time.Second) //TODO: delete after waiting for resource creation is implemented
		toUpdate, err := client.Resource(resourceSchema).Namespace(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateTo.SetResourceVersion(toUpdate.GetResourceVersion())
		_, err = client.Resource(resourceSchema).Namespace(namespace).Update(&updateTo, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		return nil
	}, m.RetryOptions...))
}

//DeleteResource deletes a given k8s resource
func (m *Manager) DeleteResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) {
	panicOnErr(retry.Do(func() error {
		deletePolicy := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		if err := client.Resource(resourceSchema).Namespace(namespace).Delete(resourceName, deleteOptions); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
		}
		return nil
	}, m.RetryOptions...))
}

//GetResource returns chosed k8s object
func (m *Manager) GetResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) (*unstructured.Unstructured, error) {
	var res *unstructured.Unstructured
	err := retry.Do(
		func() error {
			var err error
			res, err = client.Resource(resourceSchema).Namespace(namespace).Get(resourceName, metav1.GetOptions{})
			if err != nil {
				log.Printf("Error: %+v", err)
				return err
			}
			return nil
		}, m.RetryOptions...)
	if err != nil {
		log.Panicf("Error: %+v", err)
		return nil, err
	}
	log.Printf("Resource found: %+v", res.GetName())
	return res, nil
}

//GetStatus do a GetResource and extract status field
func (m *Manager) GetStatus(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) (map[string]interface{}, error) {
	obj, err := m.GetResource(client, resourceSchema, namespace, resourceName)
	if err != nil {
		log.Panicf("Error: %+v", err)
		return nil, err
	}
	status, found, err := unstructured.NestedMap(obj.Object, "status")
	if err != nil || !found {
		log.Panicf("Error: Could not retrive status, or status not found:\n %+v", err)
		return nil, err
	}
	return status, nil
}

func panicOnErr(err error) {
	if err != nil {
		log.Panicf("Error: %+v", err)
	}
}
