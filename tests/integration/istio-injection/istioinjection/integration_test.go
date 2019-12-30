package istioinjection

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIstioInjection(t *testing.T) {
	cases := []struct {
		description                   string
		disableInjectionForNamespace  bool
		disableInjectionForDeployment bool
		out                           int
	}{
		{"no injection flag for both namespace and deployment", false, false, 2},
		{"deployment injection is disabled", false, true, 1},
		{"namespace injection is disabled", true, false, 1},
		{"namespace and deployment injections are disabled", true, true, 1},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			disableInjectionForNamespace(c.disableInjectionForNamespace)
			testID := generateTestID(8)
			createDeployment(testID, c.disableInjectionForDeployment)
			defer deleteDeployment(testID)
			time.Sleep(2 * time.Second)

			pods, _ := k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=" + testID})
			pod := pods.Items[0]
			numberOfContainers := len(pod.Spec.Containers)

			if numberOfContainers != c.out {
				t.Errorf("pod has %d containers, but wanted %d containers", numberOfContainers, c.out)
			}
		})
	}
}
