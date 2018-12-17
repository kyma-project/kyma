package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	numRequest     = 1600
	functionName   = "load-test"
	namespace      = "load-test"
	functionYaml   = "k8syaml/function.yaml"
	nsYaml         = "k8syaml/ns.yaml"
	expectedOutput = "Call to the function load-test was successful!"
)

var (
	endpoint   = fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
	client     = getHttpClient(true)
	slack      *Slack
	testResult *TestResult
	timeout    = time.After(time.Duration(5) * time.Minute)
	stopping   = false
	mutex      sync.RWMutex
)

type Slack struct {
	SlackEndpoint string
	SlackChannel  string
}
type TestResult struct {
	sync.RWMutex
	resultMessage         string
	errorResponse         string
	errorRequest          string
	numFailedRequests     int
	numSuccessfulRequests int
	totalRequestsDone     int
	totalRequests         int
}

func init() {
	cleanup()
	log.Printf("Creating namespace: %s", namespace)
	createNS()
	log.Printf("Deploying function: %s", functionName)
	deployFun()
	log.Printf("Verifying correct function output for %s", functionName)
	log.Printf("HTTP endpoint for the function: %v", endpoint)
	ensureOutputIsCorrect()
	slack = NewSlack()
	testResult = NewTestResult()
}

func main() {
	log.Println("Starting Horizontal Pod Autoscaler test for functions")
	numCPUs := runtime.GOMAXPROCS(runtime.NumCPU())
	log.Println("Number of logical CPUs: ", runtime.NumCPU())
	start := time.Now()
	tick := time.Tick(1 * time.Second)
	calculateExecutionTime()
	premature := true
	done := false

	respCh := make(chan string)
	doneCh := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(numCPUs)
	testResult.totalRequests = numCPUs * numRequest
	log.Println("Number of requests to be done: ", testResult.totalRequests)
	for c := 0; c < numCPUs; c++ {
		go func() {
			defer wg.Done()
			for r := 0; r < numRequest; r++ {
				mutex.RLock()
				if stopping {
					break
				}
				mutex.RUnlock()
				makeHttpRequest(respCh)
			}
		}()
	}
	go func() {
		for resp := range respCh {
			log.Println(resp)
		}
	}()
	go func() {
		wg.Wait()
		close(respCh)
		doneCh <- true
	}()

	for {
		select {
		case <-timeout:
			mutex.Lock()
			stopping = true
			mutex.Unlock()
			done = true
		case <-tick:
			//Processing the requests
		case <-doneCh:
			log.Println("All requests were processed!")
			premature = false
			done = true
			break
		}
		if done {
			break
		}
	}
	closingTest(start, premature)

}

func makeHttpRequest(respCh chan<- string) {
	testResult.Lock()
	defer testResult.Unlock()
	testResult.totalRequestsDone++
	start := time.Now()
	testID := randomString(8)
	resp, err := http.Post(endpoint, "text/plain", bytes.NewBuffer([]byte(testID)))
	secs := time.Since(start).Seconds()
	if err != nil {
		testResult.errorRequest = fmt.Sprintf("TestId: [%v] %.2f secs elapsed with error on response [ERROR] %v", testID, secs, err)
		respCh <- testResult.errorRequest
		testResult.numFailedRequests++
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		testResult.errorResponse = fmt.Sprintf("TestId: [%v] %.2f secs elapsed with error : unable to read response [ERROR] %v", testID, secs, err)
		respCh <- testResult.errorResponse
		testResult.numFailedRequests++
		return
	}
	if resp.StatusCode != http.StatusOK {
		testResult.errorResponse = fmt.Sprintf("TestId: [%v] %.2f secs elapsed with no 200 response: response code: %v endpoint: %s", testID, secs, resp.StatusCode, endpoint)
		respCh <- testResult.errorResponse
		testResult.numFailedRequests++
		return
	}
	respCh <- fmt.Sprintf("TestId: [%v] Response code: HTTP %v, Response: %s, Response time: %.2f secs,  endpoint: %s", testID, resp.StatusCode, string([]byte(body)), secs, endpoint)
	testResult.numSuccessfulRequests++
}

func closingTest(start time.Time, premature bool) {
	checkFunctionAutoscaled(premature)
	slack.sendNotificationtoSlackChannel(testResult)
	log.Printf("%.2fm elapsed\n", time.Since(start).Minutes())
	cleanup()
	if premature {
		log.Fatalf("Load test timed out!")
	} else {
		log.Println("Horizontal Pod Autoscaler test for functions ends!")
	}
}

func calculateExecutionTime() {
	execTimeout := os.Getenv("LOAD_TEST_EXECUTION_TIMEOUT")
	if len(execTimeout) > 0 {
		executionTimeOut, err := strconv.Atoi(execTimeout)
		if err != nil {
			log.Printf("error on getting env variable for LOAD_TEST_EXECUTION_TIMEOUT: %v", execTimeout)
			log.Printf("current execution timeout %v", executionTimeOut)
		}
		timeout = time.After(time.Duration(executionTimeOut) * time.Minute)
		log.Printf("Configured %v minute(s) timeout for the execution of load-test.", executionTimeOut)
	}
}

func createNS() {
	stdoutStderr, err := deployK8s(nsYaml)
	if err != nil {
		log.Fatal("unable to create namespace ", namespace, ":\n", string(stdoutStderr))
	}
}

func deployFun() {
	stdoutStderr, err := deployK8s(functionYaml)
	if err != nil {
		log.Fatal("unable to deploy function ", functionName, ":\n", string(stdoutStderr))
	}
	log.Printf("Verifying that function %s is correctly deployed.\n", functionName)
	ensureFunctionIsRunning()
}

func ensureFunctionIsRunning() {
	timeoutFunc := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeoutFunc:
			cmd := exec.Command("kubectl", "-n", namespace, "describe", "pod", "-l", "function="+functionName)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatalf("Timed out waiting for: %v function pod to be running. Because of following error: %v ", functionName, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName, "-ojsonpath={range .items[*]}{.status.phase}{end}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error while fetching the status phase of the function pod when verifying function is running: %v", string(stdoutStderr))
			}
			functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName, "-ojsonpath={.items[*].metadata.name}")
			functionPodName, err := functionPodsCmd.CombinedOutput()
			if err != nil {
				log.Printf("Error in fetching function pod when verifying function is running: %v", string(functionPodName))
			}
			hpaOutput, err := checkFunctionHpa()
			if err != nil {
				log.Printf("Error in fetching function hpa when verifying function is running: %v", err)
			}
			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				log.Printf("Pods: %v: is running!", string(functionPodName))
				log.Printf("HPA is: %v", string(hpaOutput))
				return
			}
		}
	}
}

func ensureOutputIsCorrect() {
	timeoutOutput := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeoutOutput:
			log.Fatalf("Function is not ready: Timed out: Test HPA failed!")
		case <-tick:
			resp, err := client.Get(endpoint)
			if err != nil {
				log.Printf("Unable to call host: %v : Error: %v", endpoint, err)
			} else {
				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatalf("Unable to get response: %v", err)
					}
					log.Printf("Response from function: %v\n", string(bodyBytes))
					functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName, "-ojsonpath={.items[*].metadata.name}")
					functionPodName, err := functionPodsCmd.CombinedOutput()
					if err != nil {
						log.Printf("Error in fetch function pod when verifying correct output: %v", string(functionPodName))
					}
					if strings.Contains(string(bodyBytes), expectedOutput) {
						log.Printf("Response contains output: %v == %v", string(bodyBytes), expectedOutput)
						log.Printf("Name of the successful pod is: %v", string(functionPodName))
						return
					}
					log.Printf("Name of the failed pod is: %v", string(functionPodName))
					log.Fatalf("Response is not equal to expected output:\nResponse: %v\nExpected: %v", string(bodyBytes), expectedOutput)
				} else {
					log.Printf("Response from function: %v", resp.StatusCode)
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Printf("Unable to get response: %v", err)
					}
					log.Printf("Response body is: %v", string(bodyBytes))
				}
			}
		}
	}
}

func deployK8s(yamlFile string) (string, error) {
	cmd := exec.Command("kubectl", "create", "-f", yamlFile, "-n", namespace)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	return output, err
}

const lettersAndNums = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = lettersAndNums[rand.Intn(len(lettersAndNums))]
	}
	return string(b)
}

func getHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func cleanup() {
	log.Println("Cleaning up resources")
	deleteFun()
	deleteNamespace()
}

func deleteFun() {
	stdoutStderr, err := deleteK8s(functionYaml)
	output := string(stdoutStderr)
	if err != nil {
		log.Fatal("Unable to delete function ", functionName, ":\n", output)
	}
	timeoutFuncDelete := time.After(15 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeoutFuncDelete:
			log.Fatal("Timed out waiting for ", functionName, " pod to be deleted\n")
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName)
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(stdoutStderr), "No resources found") {
				log.Printf("All functions cleaned up!")
				return
			}
		}
	}
}

func deleteNamespace() {
	stdoutStderr, err := deleteK8s(nsYaml)
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "not found") && !strings.Contains(output, "The system is ensuring all content is removed from this namespace") {
		log.Fatal("Unable to delete namespace ", namespace, ":\n", output)
	}
	timeoutNSDelete := time.After(20 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		cmd := exec.Command("kubectl", "get", "ns", namespace, "-oyaml")
		select {
		case <-timeoutNSDelete:
			cmd = exec.Command("kubectl", "get", "po,svc,deploy,function,rs,hpa", "-n", namespace, "-oyaml")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to get ns: %v\n", string(stdoutStderr))
			}
			log.Printf("Current state of the ns: %s is:\n %v", namespace, string(stdoutStderr))
			log.Fatal("Timed out waiting for namespace: ", namespace, " to be deleted\n")
		case <-tick:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil && strings.Contains(string(stdoutStderr), "not found") {
				log.Print("No load-test ns exists!")
				return
			}
		}
	}
}

func deleteK8s(yamlFile string) (string, error) {
	cmd := exec.Command("kubectl", "delete", "-f", yamlFile, "-n", namespace, "--grace-period=0", "--force", "--ignore-not-found")
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	return output, err
}

func checkFunctionHpa() ([]byte, error) {
	functionHpaCmd := exec.Command("kubectl", "-n", namespace, "get", "hpa", "-l", "function="+functionName, "-oyaml")
	hpaOutput, err := functionHpaCmd.CombinedOutput()
	return hpaOutput, err
}

func checkFunctionAutoscaled(premature bool) {
	testResult.RLock()
	functionHpaCmd := exec.Command("kubectl", "-n", namespace, "get", "hpa", "-l", "function="+functionName, "-ojsonpath={.items[0].metadata.name} {.items[0].spec.minReplicas} {.items[0].status.currentReplicas} {.items[0].status.currentCPUUtilizationPercentage}")
	hpaOutput, err := functionHpaCmd.CombinedOutput()
	result := ""
	if err != nil {
		testResult.resultMessage = fmt.Sprintf("Error in fetching function HPA: %v \n", string(hpaOutput))
		log.Printf(testResult.resultMessage)
	} else {
		if premature {
			result = "@channel HPA test timed out!"
		} else {
			result = "Autoscaling of functions was successful!"
		}
		status := strings.Split(strings.TrimSpace(string(hpaOutput)), " ")
		minReplicas, err := strconv.Atoi(status[1])
		if err != nil {
			minReplicas = 0
		}
		minReplicasStatus := fmt.Sprintf("Minimum number of replicas: %v", minReplicas)
		currentReplicas, err := strconv.Atoi(status[2])
		if err != nil {
			currentReplicas = 0
		}
		currentReplicasStatus := fmt.Sprintf("Current number of replicas: %v", currentReplicas)

		cpuStatus := fmt.Sprintf("CPU utilization: %v%s", 0, "%")
		if len(status) == 4 {
			currentCPUUtilizationPercentage, err := strconv.Atoi(status[3])
			if err != nil {
				currentCPUUtilizationPercentage = 0
			}
			cpuStatus = fmt.Sprintf("CPU utilization: %v%s", currentCPUUtilizationPercentage, "%")
		}

		if currentReplicas > minReplicas {
			result = "Function autoscale succeeded"
		}
		finalStatus := fmt.Sprintf("Test HPA final status: %s \n%s \n%s \n%s\n", result, minReplicasStatus, currentReplicasStatus, cpuStatus)
		testResult.resultMessage = finalStatus

		if testResult.totalRequests > 0 {
			totalRequests := fmt.Sprintf("Total number of requests: %v \n", testResult.totalRequests)
			testResult.resultMessage = fmt.Sprintf("%s %s\n", testResult.resultMessage, strings.TrimSpace(totalRequests))
		}

		if testResult.totalRequestsDone > 0 {
			totalRequestsDone := fmt.Sprintf("Total number of requests done: %v \n", testResult.totalRequestsDone)
			testResult.resultMessage = fmt.Sprintf("%s %s\n", testResult.resultMessage, strings.TrimSpace(totalRequestsDone))
		}

		if testResult.numSuccessfulRequests > 0 {
			numSuccessfulRequests := fmt.Sprintf("Successful requests: %v \n", testResult.numSuccessfulRequests)
			testResult.resultMessage = fmt.Sprintf("%s %s\n", testResult.resultMessage, strings.TrimSpace(numSuccessfulRequests))
		}

		if testResult.numFailedRequests > 0 {
			numFailedRequests := fmt.Sprintf("Failed resquests: %v \n", testResult.numFailedRequests)
			testResult.resultMessage = fmt.Sprintf("%s %s\n", testResult.resultMessage, strings.TrimSpace(numFailedRequests))
			testResult.totalRequests = testResult.totalRequests + testResult.numFailedRequests
		}

		log.Println(testResult.resultMessage)
		testResult.RUnlock()
	}
}

func (slack *Slack) sendNotificationtoSlackChannel(testResult *TestResult) {
	textMessage := fmt.Sprintf(`{"channel": "%v", "text":"%v"}"`, slack.SlackChannel, testResult.resultMessage)
	var jsonStr = []byte(textMessage)
	req, err := http.NewRequest("POST", slack.SlackEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to send slack notification to endpoint: %v : Error: %v", slack.SlackChannel, err)
	}
	defer resp.Body.Close()
	log.Println("Slack response status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("Slack response response body:", string(body))
}

func NewSlack() *Slack {
	slackEndpoint := os.Getenv("LOAD_TEST_SLACK_ENDPOINT")
	if len(slackEndpoint) == 0 {
		log.Fatalln("No slack endpoint provided!")
	}

	apiToken := os.Getenv("LOAD_TEST_SLACK_TOKEN")
	if len(apiToken) == 0 {
		log.Fatalln("No slack api token provided!")
	}
	sUrl := fmt.Sprintf("%s%s", slackEndpoint, apiToken)

	sChannel := os.Getenv("LOAD_TEST_SLACK_CHANNEL")
	if len(sChannel) == 0 {
		log.Fatalln("No slack channel provided!")
	}
	s := &Slack{sUrl, sChannel}

	log.Printf("Slack configuration: %v", s)
	return s
}

func NewTestResult() *TestResult {
	t := &TestResult{}
	return t
}
