package logstream

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitForDummyPodToRun waits until the dummy pod is running
func WaitForDummyPodToRun(namespace string, coreInterface kubernetes.Interface) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			log.Fatal("Timed out while waiting for test-counter-pod to be Running!")
		case <-tick:
			pod, err := coreInterface.CoreV1().Pods(namespace).Get("test-counter-pod", metav1.GetOptions{})
			if err != nil {
				log.Fatalf("Unable to get pod: %v", err)
			}
			if pod.Status.Phase == corev1.PodRunning {
				log.Println("test-counter-pod is running!")
				return
			}
		}
	}
}

// Test querys loki api with the given label key-value pair and checks that the logs of the dummy pod are present
func Test(labelKey string, labelValue string, authHeader string, startTime int64) {
	timeout := time.After(1 * time.Minute)
	tick := time.Tick(1 * time.Second)
	lokiURL := "http://logging-loki.kyma-system:3100/api/prom/query"
	query := fmt.Sprintf("query={%s=\"%s\"}", labelKey, labelValue)
	startTimeParam := fmt.Sprintf("start=%s", strconv.FormatInt(startTime, 10))
	for {
		select {
		case <-timeout:
			log.Fatalf("The string 'logTest-' is not present in logs when using the following query: %s", query)
		case <-tick:
			cmd := exec.Command("curl", "-v", "-G", "-s", lokiURL, "--data-urlencode", query, "--data-urlencode", startTimeParam, "-H", authHeader)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error in HTTP GET to %s: %v\n%s", lokiURL, err, string(stdoutStderr))
			}

			var testDataRegex = regexp.MustCompile(`logTest-`)
			submatches := testDataRegex.FindStringSubmatch(string(stdoutStderr))
			if submatches != nil {
				log.Printf("The string 'logTest-' is present in logs when using the following query: %s", query)
				return
			}
		}
	}
}

// Cleanup terminates the dummy pod
func Cleanup(namespace string, coreInterface kubernetes.Interface) {
	gracePeriod := int64(0)
	deleteOptions := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	listOptions := metav1.ListOptions{LabelSelector: "app=test-counter-pod"}
	err := coreInterface.CoreV1().Pods(namespace).DeleteCollection(&deleteOptions, listOptions)
	if err != nil {
		log.Fatalf("Unable to delete test-counter-pod: %v", err)
	}
	log.Println("Cleanup is successful!")
}
