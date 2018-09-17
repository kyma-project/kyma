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

const expectedLogSpout = 1
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

func testPodsAreReady() {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)

	for {
		actualLogSpout := 0
		actualOKLog := 0

		select {
		case <-timeout:
			log.Println("Test if all the OKLog pods are up and running: result: Timed out!!")
			if expectedLogSpout != actualLogSpout {
				log.Fatalf("Expected 'Logspout' pods healthy is %d but got %d instances", expectedLogSpout, actualLogSpout)
			}
			if expectedOKLog != actualOKLog {
				log.Fatalf("Expected 'OKLog' pods healthy is %d but got %d instances", expectedOKLog, actualOKLog)
			}
		case <-tick:
			cmd := exec.Command("kubectl", "get", "pods", "-l", "component in (oklog, logspout)", "-n", namespace, "--no-headers")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error while kubectl get: %s ", err)
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
				log.Println("Test pods status: All OKLog pods are ready!!")
				return
			}
			log.Println("Waiting for OKLog pods to be READY!!")
		}
	}
}

func testOKLogUI() {
	resp, err := http.Get("http://core-logging-oklog-0.core-logging-oklog:7650/ui/")

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

func testLogSpout() {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)
	var testDataRegex = regexp.MustCompile(`(?m)core-logging-oklog\.kyma-system:7651*`)

	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "logs", "-l", "component=logspout")
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatal("Timed out getting the correct log for pod logspout", ":\n", string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "logs", "-l", "component=logspout")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatal("Unable to obtain function log for pod logspout", ":\n", string(stdoutStderr))
			}

			submatches := testDataRegex.FindStringSubmatch(string(stdoutStderr))
			if submatches != nil {
				log.Printf("Test Check LogSpout has Passed. OKLog Service is set in logspout: %v", submatches)
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
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			log.Println("Test LogStreaming: result: Timed out!!")
			cmd := exec.Command("kubectl", "get", "pods", "-l", "component=test-counter-pod", "-n", namespace, "--no-headers")
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

	res, err := c.Get("http://core-logging-oklog-0.core-logging-oklog:7650/store/stream?q=oklogTest-")
	if err != nil {
		log.Fatalf("Error in HTTP GET to http://core-logging-oklog-0.core-logging-oklog:7650/store/stream?q=oklogTest-: %v\n", err)
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
