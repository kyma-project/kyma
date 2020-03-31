package logstream

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// WaitForDummyPodToRun waits until the dummy pod is running
func WaitForDummyPodToRun(namespace string) {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			log.Println("Test LogStreaming: result: Timed out!!")
			cmd := exec.Command("kubectl", "describe", "pods", "-l", "component=test-counter-pod", "-n", namespace)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatal("Test LogStreaming: result: Timed out!! Current state is", ":\n", string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "test-counter-pod", "-ojsonpath={.status.phase}")
			stdoutStderr, err := cmd.CombinedOutput()

			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				log.Println("test-counter-pod is running!")
				return
			}
			log.Println("Waiting for the test-counter-pod to be Running!")
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
func Cleanup(namespace string) {
	cmd := exec.Command("kubectl", "-n", namespace, "delete", "pod", "-l", "app=test-counter-pod", "--force", "--grace-period=0")
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "NotFound") {
		log.Fatalf("Unable to delete test-counter-pod: %s", output)
	}
	log.Println("Cleanup is successful!")
}
