package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const expectedOKLog = 1
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
	cmd := exec.Command("kubectl", "get", "nodes")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while kubectl get nodes: %v", string(stdoutStderr))
	}
	outputArr := strings.Split(string(stdoutStderr), "\n")
	return len(outputArr) - 2
}

func testPodsAreReady() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(5 * time.Second)
	expectedLogSpout := getNumberOfNodes()
	for {
		actualLogSpout := 0
		actualOKLog := 0

		select {
		case <-timeout:
			if expectedLogSpout != actualLogSpout {
				log.Printf("Expected 'Logspout' pods healthy is %d but got %d instances", expectedLogSpout, actualLogSpout)
				cmd := exec.Command("kubectl", "describe", "pods", "-l", "component=logspout", "-n", namespace)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Error while kubectl describe: %s ", string(stdoutStderr))
				}
				log.Printf("Existing pods for logspout:\n%s\n", string(stdoutStderr))
			}
			if expectedOKLog != actualOKLog {
				log.Printf("Expected 'OKLog' pods healthy is %d but got %d instances", expectedOKLog, actualOKLog)
				cmd := exec.Command("kubectl", "describe", "pods", "-l", "component=oklog", "-n", namespace)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Error while kubectl describe: %s ", string(stdoutStderr))
				}
				log.Printf("Existing pods for oklog:\n%s\n", string(stdoutStderr))
			}
			log.Fatalf("Test if all the OKLog/Logspout pods are up and running: result: Timed out!!")
		case <-tick:
			cmd := exec.Command("kubectl", "get", "pods", "-l", "component in (oklog, logspout)", "-n", namespace, "--no-headers")
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
						case strings.Contains(podName, "logspout"):
							actualLogSpout++

						case strings.Contains(podName, "oklog"):
							actualOKLog++
						}
					}
				}
			}
			if expectedLogSpout == actualLogSpout && expectedOKLog == actualOKLog {
				log.Println("Test pods status: All OKLog/LogSpout pods are ready!!")
				return
			}
			log.Println("Waiting for OKLog/Logspout pods to be READY!!")
		}
	}
}

func testOKLogUI() {
	resp, err := http.Get("http://logging-oklog-0.logging-oklog:7650/ui/")

	if err != nil {
		log.Fatalf("Test Check OKLogUI Failed: error is: %v and response is: %v", err, resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Test Check OKLogUI Failed: Unable to fetch UI. Response code is: %v and response test is: %v", resp.StatusCode, string(body))
	}
	log.Printf("Test Check OKLogUI Passed. Response code is: %v", resp.StatusCode)
}

func getLogSpoutPods() []string {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", "component=logspout", "-ojsonpath={range .items[*]}{.metadata.name}\n{end}")

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while getting all logspout pods: %v", string(stdoutStderr))
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

func testLogSpout() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	var testDataRegex = regexp.MustCompile(`(?m)logging-oklog\.kyma-system:7651*`)
	pods := getLogSpoutPods()
	log.Println("LogSpout pods are: ", pods)
	for {
		select {
		case <-timeout:
			for _, pod := range pods {
				cmd := exec.Command("kubectl", "-n", namespace, "log", pod)
				stdoutStderr, _ := cmd.CombinedOutput()
				log.Printf("Logs for pod %s:\n%s", pod, string(stdoutStderr))
			}
			log.Fatalf("Timed out getting the correct logs for Logspout pods")
		case <-tick:
			matchesCount := 0
			for _, pod := range pods {
				cmd := exec.Command("kubectl", "-n", namespace, "log", pod)
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
				log.Printf("Test Check LogSpout passed. OKLog Service is set in all %d logsSpout pods", len(pods))
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
				log.Printf("test-counter-pod is running!")
				return
			}
			log.Printf("Waiting for the test-counter-pod to be Running!")
		}
	}
}

func testLogs() {
	c := &http.Client{
		Timeout: 45 * time.Second,
	}

	res, err := c.Get("http://logging-oklog-0.logging-oklog:7650/store/stream?q=oklogTest-")
	if err != nil {
		log.Fatalf("Error in HTTP GET to http://logging-oklog-0.logging-oklog:7650/store/stream?q=oklogTest-: %v\n", err)
	}
	defer res.Body.Close()
	var testDataRegex = regexp.MustCompile(`(?m)oklogTest-*`)

	reader := bufio.NewReader(res.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalf("Error in reading from log stream: %v", err)
			return
		}
		submatches := testDataRegex.FindStringSubmatch(string(line))
		if submatches != nil {
			log.Printf("The string 'oklogtest-' is present in logs: %v", string(line))
			return
		}
	}
}

func testLogStream() {
	log.Println("Deploying test-counter-pod Pod")
	deployDummyPod()
	waitForDummyPodToRun()
	testLogs()
	log.Println("Test Logging Succeeded!")
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
	log.Println("Test if all the OKLog pods are ready")
	testPodsAreReady()
	log.Println("Test if all the OKLog UI is reachable")
	testOKLogUI()
	log.Println("Test if LogSpout is able to find OKLog")
	testLogSpout()
	log.Println("Test if logs from a dummy Pod are streamed by OKLog")
	testLogStream()
	cleanup()
}
