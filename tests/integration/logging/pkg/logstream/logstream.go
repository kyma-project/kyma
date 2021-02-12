package logstream

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/kyma/tests/integration/logging/pkg/request"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//PodSpec represents a test logging pod configuration
type PodSpec struct {
	PodName       string
	ContainerName string
	Namespace     string
	LogPrefix     string
}

// DeployDummyPod deploys a pod which keeps logging a counter
func DeployDummyPod(spec PodSpec, coreInterface kubernetes.Interface) error {
	labels := map[string]string{
		"app": spec.PodName,
	}
	args := []string{
		"sh",
		"-c",
		fmt.Sprintf("let i=1; while true; do echo \"$i: %s-$(date)\"; let i++; sleep 2; done", spec.LogPrefix),
	}

	_, err := coreInterface.CoreV1().Pods(spec.Namespace).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   spec.PodName,
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  spec.ContainerName,
					Image: "alpine:3.8",
					Args:  args,
				},
			},
		},
	})
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", spec.PodName)
	}
	return nil
}

// WaitForDummyPodToRun waits until the dummy pod is running
func WaitForDummyPodToRun(spec PodSpec, coreInterface kubernetes.Interface) error {
	timeout := time.After(2 * time.Minute)
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf("timed out while waiting for %s to be Running!", spec.PodName)
		case <-tick.C:
			pod, err := coreInterface.CoreV1().Pods(spec.Namespace).Get(spec.PodName, metav1.GetOptions{})
			if err != nil {
				return errors.Wrapf(err, "cannot get %s", spec.PodName)
			}
			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
		}
	}
}

// Test querys loki api with the given label key-value pair and checks that the logs of the dummy pod are present
func Test(domain, authHeader, logPrefix string, labelsToSelect map[string]string, httpClient *http.Client) error {
	start := time.Now().UnixNano()
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	query := logQLEncode(labelsToSelect)
	lokiURL := fmt.Sprintf(`https://loki.%s/api/prom/query?query=%s&start=%d`, domain, query, start)
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf(`the string "%s" is not present in logs when using the following query: %s`, logPrefix, query)
		case <-tick.C:
			respStatus, respBody, err := request.DoGet(httpClient, lokiURL, authHeader)
			if err != nil {
				return errors.Wrap(err, "cannot query loki for logs")
			}
			if respStatus != http.StatusOK {
				return errors.Errorf("error in HTTP GET to %s.\nStatus Code: %d\nResponse: %s", lokiURL, respStatus, respBody)
			}
			var testDataRegex = regexp.MustCompile(logPrefix)
			submatches := testDataRegex.FindStringSubmatch(respBody)
			if submatches != nil {
				return nil
			}
		}
	}
}

func logQLEncode(labelsToSelect map[string]string) string {
	var sb strings.Builder
	sb.WriteRune('{')
	keyIndex := 0
	for key, value := range labelsToSelect {
		sb.WriteString(fmt.Sprintf(`%s="%s"`, key, value))

		if keyIndex < len(labelsToSelect)-1 {
			sb.WriteRune(',')
		}

		keyIndex++
	}
	sb.WriteRune('}')
	return sb.String()
}

// Cleanup terminates the dummy pod
func Cleanup(spec PodSpec, coreInterface kubernetes.Interface) error {
	gracePeriod := int64(0)
	deleteOptions := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", spec.PodName)}
	err := coreInterface.CoreV1().Pods(spec.Namespace).DeleteCollection(&deleteOptions, listOptions)
	if err != nil {
		return errors.Wrapf(err, "cannot delete %s", spec.PodName)
	}
	return nil
}
