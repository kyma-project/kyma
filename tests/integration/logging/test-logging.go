package main

import (
	"fmt"
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
	linesToRemove := 1
	if strings.Contains(string(stdoutStderr), "master") && !strings.Contains(string(stdoutStderr), "minikube") && !strings.Contains(string(stdoutStderr), "control-plane") {
		linesToRemove++
	}
	outputArr := strings.Split(string(stdoutStderr), "\n")

	return len(outputArr) - linesToRemove
}

func testPodsAreReady() {
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

func testFluentBit() {
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

func deployDummyPod() {
	cmd := exec.Command("kubectl", "-n", namespace, "create", "-f", yamlFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy:\n", string(stdoutStderr))
	}
}

func waitForDummyPodToRun() {
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

func setAuthHeader() string {
	token := getJWT()
	authHeader := fmt.Sprintf("Authorization: Bearer %s", token)
	return authHeader
}

func testLogs(labelKey string, labelValue string, authHeader string) {
	timeout := time.After(1 * time.Minute)
	tick := time.Tick(1 * time.Second)
	query := fmt.Sprintf("query={%s=\"%s\"}", labelKey, labelValue)

	for {
		select {
		case <-timeout:
			log.Fatalf("The string 'logTest-' is not present in logs when using the following query: %s", query)
		case <-tick:
			cmd := exec.Command("curl", "-G", "-s", "http://logging-loki:3100/api/prom/query", "--data-urlencode", query, "-H", authHeader)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error in HTTP GET to http://logging-loki:3100/api/prom/query: %v\n%s", err, string(stdoutStderr))
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

func testLogStream() {
	log.Println("Deploying test-counter-pod Pod")
	deployDummyPod()
	waitForDummyPodToRun()
	authHeader := setAuthHeader()
	testLogs("container", "count", authHeader)
	testLogs("app", "test-counter-pod", authHeader)
	testLogs("namespace", "kyma-system", authHeader)
	log.Println("Test Logging Succeeded!")
}

func cleanup() {
	cmd := exec.Command("kubectl", "-n", namespace, "delete", "-f", yamlFile, "--force", "--grace-period=0")
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "NotFound") {
		log.Fatalf("Unable to delete test-counter-pod: %s", output)
	}
	log.Println("Cleanup is successful!")
}

func main() {
	log.Println("Starting logging test")
	cleanup()
	log.Println("Test if all the Loki pods are ready")
	testPodsAreReady()
	log.Println("Test if Fluent Bit is able to find Loki")
	testFluentBit()
	log.Println("Test if logs from a dummy Pod are streamed by Loki")
	testLogStream()
	cleanup()
}
