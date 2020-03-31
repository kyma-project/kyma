package loki

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const expectedLoki = 1
const namespace = "kyma-system"

func getPodStatus(stdout string) (podName string, isReady bool) {
	isReady = false
	stdoutArr := regexp.MustCompile("( )+").Split(stdout, -1)
	podName = stdoutArr[0]
	readyCount := strings.Split(stdoutArr[1], "/")
	status := stdoutArr[2]
	if strings.ToUpper(status) == "RUNNING" && readyCount[0] == readyCount[1] {
		isReady = true
	}
	return
}

func getNumberOfNodes() int {
	cmd := exec.Command("kubectl", "get", "nodes", "--no-headers")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while kubectl get nodes: %v", string(stdoutStderr))
	}
	linesToRemove := 1
	if strings.Contains(string(stdoutStderr), "master") && !strings.Contains(string(stdoutStderr), "minikube") && !strings.Contains(string(stdoutStderr), "control-plane") {
		linesToRemove++
	}
	outputArr := strings.Split(string(stdoutStderr), "\n")

	return len(outputArr) - linesToRemove
}

// TestPodsAreReady checks if all the Loki/Fluent Bit pods are ready
func TestPodsAreReady() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(5 * time.Second)
	expectedFluentbits := getNumberOfNodes()
	for {
		actualFluentbits := 0
		actualLoki := 0

		select {
		case <-timeout:
			if expectedFluentbits != actualFluentbits {
				log.Printf("Expected 'fluent-bit' pods healthy is %d but got %d instances", expectedFluentbits, actualFluentbits)
				cmd := exec.Command("kubectl", "describe", "pods", "-l", "app=fluent-bit", "-n", namespace)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Error while kubectl describe: %s ", string(stdoutStderr))
				}
				log.Printf("Existing pods for fluent-bit:\n%s", string(stdoutStderr))
			}
			if expectedLoki != actualLoki {
				log.Printf("Expected 'Loki' pods healthy is %d but got %d instances", expectedLoki, actualLoki)
				cmd := exec.Command("kubectl", "describe", "pods", "-l", "app=loki", "-n", namespace)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Error while kubectl describe: %s ", string(stdoutStderr))
				}
				log.Printf("Existing pods for loki:\n%s", string(stdoutStderr))
			}
			log.Fatal("Test if all the Loki/Fluent Bit pods are up and running: result: Timed out!!")
		case <-tick:
			cmd := exec.Command("kubectl", "get", "pods", "-l", "app in (loki, fluent-bit)", "-n", namespace, "--no-headers")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error while kubectl get: %s ", string(stdoutStderr))
			}
			outputArr := strings.Split(string(stdoutStderr), "\n")

			for index := range outputArr {
				if len(outputArr[index]) != 0 {
					podName, isReady := getPodStatus(string(outputArr[index]))
					if isReady {
						switch true {
						case strings.Contains(podName, "fluent-bit"):
							actualFluentbits++

						case strings.Contains(podName, "logging") && !strings.Contains(podName, "fluent-bit"):
							actualLoki++
						}
					}
				}
			}
			if expectedFluentbits == actualFluentbits && expectedLoki == actualLoki {
				log.Println("Test pods status: All Loki/Fluent Bit pods are ready!!")
				return
			}
			log.Println("Waiting for Loki/Fluent Bit pods to be READY!!")
		}
	}
}
