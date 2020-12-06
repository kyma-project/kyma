package logstream

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeployDummyPod deploys a pod which keeps logging a counter
func DeployDummyPod(namespace string, coreInterface kubernetes.Interface) error {
	labels := map[string]string{
		"app": "test-counter-pod",
	}
	args := []string{"sh", "-c", "let i=1; while true; do echo \"$i: logTest-$(date)\"; let i++; sleep 2; done"}

	_, err := coreInterface.CoreV1().Pods(namespace).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-counter-pod",
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  "count",
					Image: "alpine:3.8",
					Args:  args,
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "cannot create test-counter-pod")
	}
	return nil
}

// WaitForDummyPodToRun waits until the dummy pod is running
func WaitForDummyPodToRun(namespace string, coreInterface kubernetes.Interface) error {
	timeout := time.After(2 * time.Minute)
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf("timed out while waiting for test-counter-pod to be Running!")
		case <-tick.C:
			pod, err := coreInterface.CoreV1().Pods(namespace).Get("test-counter-pod", metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "cannot get test-counter-pod")
			}
			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
		}
	}
}

// Test querys loki api with the given label key-value pair and checks that the logs of the dummy pod are present
func Test(domain string, labelKey string, labelValue string, authHeader string, httpClient *http.Client) error {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	currentTimeUnixNano := time.Now().UnixNano()
	startTimeParam := strconv.FormatInt(currentTimeUnixNano, 10)
	lokiURL := fmt.Sprintf(`https://loki.%s/api/prom/query?query={%s="%s"}&start=%s`, domain, labelKey, labelValue, startTimeParam)
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf(`the string "logTest-" is not present in logs when using the following query: {%s="%s"}`, labelKey, labelValue)
		case <-tick.C:
			respBody, err := doGet(httpClient, lokiURL, authHeader)
			if err != nil {
				return errors.Wrap(err, "cannot query loki for logs")
			}
			var testDataRegex = regexp.MustCompile(`logTest-`)
			submatches := testDataRegex.FindStringSubmatch(respBody)
			if submatches != nil {
				return nil
			}
		}
	}
}

// Cleanup terminates the dummy pod
func Cleanup(namespace string, coreInterface kubernetes.Interface) error {
	gracePeriod := int64(0)
	deleteOptions := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	listOptions := metav1.ListOptions{LabelSelector: "app=test-counter-pod"}
	err := coreInterface.CoreV1().Pods(namespace).DeleteCollection(&deleteOptions, listOptions)
	if err != nil {
		return errors.Wrap(err, "cannot delete test-counter-pod")
	}
	return nil
}

func doGet(httpClient *http.Client, url string, authHeader string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrap(err, "cannot create a new HTTP request")
	}
	req.Header.Add("Authorization", authHeader)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "cannot send HTTP request to %s", url)
	}
	defer resp.Body.Close()
	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		return "", errors.Wrap(err, "cannot read response body")
	}
	if resp.StatusCode != 200 {
		return "", errors.Errorf("error in HTTP GET to %s.\nStatus Code: %d\nResponse: %s", url, resp.StatusCode, body.String())
	}
	return body.String(), nil
}
