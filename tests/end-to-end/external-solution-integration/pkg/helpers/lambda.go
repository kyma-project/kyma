package helpers

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	LambdaPort               = 8080
	LambdaPayload            = "payload"
	KymaIntegrationNamespace = "kyma-integration"
	DefaultBrokerName        = "default"
)

// LambdaHelper adds utilities to deal with lambdas
type LambdaHelper struct {
	pods corev1client.PodInterface
}

// NewLambdaHelper returns new LambdaHelper
func NewLambdaHelper(pods corev1client.PodInterface) *LambdaHelper {
	return &LambdaHelper{pods: pods}
}

// ListLambdaPods returns all pods for lambda
func (h *LambdaHelper) ListLambdaPods(name string) ([]v1.Pod, error) {
	labelSelector := map[string]string{
		"function":   name,
		"created-by": "kubeless",
	}
	listOptions := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelSelector).String()}
	podList, err := h.pods.List(listOptions)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

func LambdaInClusterEndpoint(name, namespace string, port int) string {
	return fmt.Sprintf("http://%s.%s:%v", name, namespace, port)
}
