package fluentbit

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const namespace = "kyma-system"

func getFluentBitPods() []string {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", "app=fluent-bit", "-ojsonpath={range .items[*]}{.metadata.name}\n{end}")

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while getting all fluent-bit pods: %s", string(stdoutStderr))
	}
	pods := strings.Split(string(stdoutStderr), "\n")
	var podsCleaned []string
	for _, pod := range pods {
		if len(strings.Trim(pod, " ")) != 0 {
			podsCleaned = append(podsCleaned, pod)
		}
	}
	return podsCleaned
}

// Test checks if Fluent Bit is able to find Loki
func Test() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	var testDataRegex = regexp.MustCompile(`(?m)logging-fluent-bit.*`)
	pods := getFluentBitPods()
	log.Println("Fluent Bit pods are: ", pods)
	for {
		select {
		case <-timeout:
			for _, pod := range pods {
				cmd := exec.Command("kubectl", "-n", namespace, "log", pod, "-c", "logging")
				stdoutStderr, _ := cmd.CombinedOutput()
				log.Printf("Logs for pod %s:\n%s", pod, string(stdoutStderr))
			}
			log.Fatal("Timed out getting the correct logs for Logspout pods")
		case <-tick:
			matchesCount := 0
			for _, pod := range pods {
				cmd := exec.Command("kubectl", "-n", namespace, "log", pod, "-c", "fluent-bit")
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Unable to obtain log for pod[%s]:\n%s", pod, string(stdoutStderr))
				}
				submatches := testDataRegex.FindStringSubmatch(string(stdoutStderr))
				if submatches != nil {
					matchesCount++
					log.Printf("Matched logs from pod: [%s]\n%v", pod, submatches)
				}
			}
			if matchesCount == len(pods) {
				log.Printf("Test Check Fluent Bit passed. Loki Service is set in all %d Fluent Bit pods", len(pods))
				return
			}
		}
	}
}
