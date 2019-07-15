package testkit

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	coreApi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
)

type LambdaHelper struct {
	pods coreClient.PodInterface
}

func NewLambdaHelper(pods coreClient.PodInterface) *LambdaHelper {
	return &LambdaHelper{pods: pods}
}

func (h *LambdaHelper) ListLambdaPods() ([]coreApi.Pod, error) {
	labelSelector := map[string]string{
		"function":   consts.AppName,
		"created-by": "kubeless",
	}
	listOptions := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelSelector).String()}
	podList, err := h.pods.List(listOptions)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}
