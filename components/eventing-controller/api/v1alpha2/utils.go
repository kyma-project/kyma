package v1alpha2

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  GroupVersion.Version,
		Group:    GroupVersion.Group,
		Resource: "subscriptions",
	}
}

func ConvertUnstructuredListToSubscriptionList(unstructuredList *unstructured.UnstructuredList) (*SubscriptionList, error) {
	subscriptionList := new(SubscriptionList)
	subscriptionListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(subscriptionListBytes, subscriptionList)
	if err != nil {
		return nil, err
	}
	return subscriptionList, nil
}
