package main

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const expectedLoki = 1
const namespace = "kyma-system"
const yamlFile = "testCounterPod.yaml"

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
	linesToRemove := 1;
	if strings.Contains(string(stdoutStderr), "master") && !strings.Contains(string(stdoutStderr), "minikube") {
		linesToRemove++
	}
	outputArr := strings.Split(string(stdoutStderr), "\n")

	return len(outputArr) - linesToRemove
}

func testPodsAreReady() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(5 * time.Second)
	expectedPromtails := getNumberOfNodes()
	for {
		actualPromtail := 0
		actualLoki := 0

		select {
		case <-timeout:
			if expectedPromtails != actualPromtail {
				log.Printf("Expected 'promtail' pods healthy is %d but got %d instances", expectedPromtails, actualPromtail)
				cmd := exec.Command("kubectl", "describe", "pods", "-l", "app=promtail", "-n", namespace)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Error while kubectl describe: %s ", string(stdoutStderr))
				}
				log.Printf("Existing pods for promtail:\n%s\n", string(stdoutStderr))
			}
			if expectedLoki != actualLoki {
				log.Printf("Expected 'Loki' pods healthy is %d but got %d instances", expectedLoki, actualLoki)
				cmd := exec.Command("kubectl", "describe", "pods", "-l", "app=loki", "-n", namespace)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Error while kubectl describe: %s ", string(stdoutStderr))
				}
				log.Printf("Existing pods for loki:\n%s\n", string(stdoutStderr))
			}
			log.Fatalf("Test if all the Loki/Promtail pods are up and running: result: Timed out!!")
		case <-tick:
			cmd := exec.Command("kubectl", "get", "pods", "-l", "app in (loki, promtail)", "-n", namespace, "--no-headers")
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
						case strings.Contains(podName, "promtail"):
							actualPromtail++

						case strings.Contains(podName, "logging") && !strings.Contains(podName, "promtail"):
							actualLoki++
						}
					}
				}
			}
			if expectedPromtails == actualPromtail && expectedLoki == actualLoki {
				log.Println("Test pods status: All Loki/Promtail pods are ready!!")
				return
			}
			log.Println("Waiting for Loki/Promtail pods to be READY!!")
		}
	}
}

func getPromtailPods() []string {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", "app=promtail", "-ojsonpath={range .items[*]}{.metadata.name}\n{end}")

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while getting all promtail pods: %v", string(stdoutStderr))
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

func testPromtail() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	var testDataRegex = regexp.MustCompile(`(?m)logging-promtail.*`)
	pods := getPromtailPods()
	log.Println("Promtail pods are: ", pods)
	for {
		select {
		case <-timeout:
			for _, pod := range pods {
				cmd := exec.Command("kubectl", "-n", namespace, "log", pod, "-c", "logging")
				stdoutStderr, _ := cmd.CombinedOutput()
				log.Printf("Logs for pod %s:\n%s", pod, string(stdoutStderr))
			}
			log.Fatalf("Timed out getting the correct logs for Logspout pods")
		case <-tick:
			matchesCount := 0
			for _, pod := range pods {
				cmd := exec.Command("kubectl", "-n", namespace, "log", pod, "-c", "promtail")
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Unable to obtain log for pod[%s]:\n%s\n", pod, string(stdoutStderr))
				}
				submatches := testDataRegex.FindStringSubmatch(string(stdoutStderr))
				if submatches != nil {
					matchesCount += 1
					log.Printf("Matched logs from pod: [%s]\n%v", pod, submatches)
				}
			}
			if matchesCount == len(pods) {
				log.Printf("Test Check Promtail passed. Loki Service is set in all %d promtail pods", len(pods))
				return
			}
		}
	}
}

func cleanup() {
	cmd := exec.Command("kubectl", "-n", namespace, "delete", "-f", yamlFile, "--force", "--grace-period=0")
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "NotFound") {
		log.Fatalf("Unable to delete test-counter-pod:%s\n", output)
	}
	log.Println("Cleanup is successful!")
}

func main() {
	log.Println("Starting logging test")
	cleanup()
	log.Println("Test if all the Loki pods are ready")
	testPodsAreReady()
	log.Println("Test if Promtail is able to find Loki")
	testPromtail()
	cleanup()
}