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
	numRequest     = 800
	functionName   = "load-test"
	namespace      = "load-test"
	functionYaml   = "k8syaml/function.yaml"
	nsYaml         = "k8syaml/ns.yaml"
	expectedOutput = "Call to the function load-test was successful!"
)

var (
	endpoint         = fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
	slackEndpoint    = "https://sap-cx.slack.com/services/hooks/jenkins-ci/gZJI7risPpW67frP3EiDrPV0"
	slackChaneel     = "#c4-xf-load-tests"
	client           = getHttpClient(true)
	slack            *Slack
	testResult       *TestResult
	executionTimeOut = 30
)

type Slack struct {
	SlackEndpoint string
	SlackChaneel  string
}
type TestResult struct {
	resultMessage         string
	errorResponse         string
	errorRequest          string
	numFailedRequests     int
	numSuccessfulRequests int
	totalRequests         int
}

func main() {
	start := time.Now()
	log.Println("Starting horizontal pod autoscaler test for functions")
	numCPUs := runtime.GOMAXPROCS(runtime.NumCPU())
	log.Println("logical CPUs: ", runtime.NumCPU())
	// Test running 30 minutes until timeout
	//timeout := time.After(1 * time.Minute)
	// a numCPUs goroutines triggered every tick
	//tick := time.Tick(1 * time.Second)
	respCh := make(chan string)
	doneCh := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(numCPUs)
	for c := 0; c < numCPUs; c++ {
		go func() {
			defer wg.Done()
			for r := 0; r < numRequest; r++ {
				if testResult.numFailedRequests > 0 {
					break
				}
				makeHttpRequest(respCh)
			}
		}()
	}
	//for {
	//
	//	select {
	//	case <-timeout:
	//		return
	//	case <-tick:
	//	}
	//}
	go func() {
		printResponse(respCh, doneCh)
	}()
	go func() {
		wg.Wait()
		close(respCh)
	}()
	<-doneCh
	checkFunctionAutoscaled()
	slack.sendNotificationtoSlackChannel(testResult)
	log.Println("Finishing horizontal pod autoscaler test for functions")
	log.Printf("%.2fs elapsed\n", time.Since(start).Minutes())
	cleanup()
}

func init() {
	cleanup()
	log.Printf("Create namespace %s \n", namespace)
	createNS()
	log.Printf("Deploying %s function \n", functionName)
	deployFun()
	log.Printf("Verifying correct function output for %s \n", functionName)
	log.Printf("Endpoint for the function: %v\n", endpoint)
	ensureOutputIsCorrect()
	slack = NewSlack()
	testResult = NewTestResult()
	execTimeout := os.Getenv("LOAD_TEST_EXECUTION_TIMEOUT")
	if len(execTimeout) > 0 {
		timeout, err := strconv.Atoi(execTimeout)
		if err != nil {
			log.Printf("Error on getting env variable for LOAD_TEST_EXECUTION_TIMEOUT: %v", execTimeout)
			log.Printf("Current execution timeout %v", executionTimeOut)
		}
		executionTimeOut = timeout
	}
	log.Printf("Scheduled execution time for load test %v minutes", executionTimeOut)
}

func createNS() {
	stdoutStderr, err := deployK8s(nsYaml)
	if err != nil {
		log.Fatal("Unable to create namespace ", namespace, ":\n", string(stdoutStderr))
	}
}

func deployFun() {
	stdoutStderr, err := deployK8s(functionYaml)
	if err != nil {
		log.Fatal("Unable to deploy function ", functionName, ":\n", string(stdoutStderr))
	}
	ensureFunctionIsRunning()
}

func ensureFunctionIsRunning() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "describe", "pod", "-l", "function="+functionName)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatalf("Timed out waiting for: %v function pod to be running. Because of following error: %v ", functionName, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName, "-ojsonpath={range .items[*]}{.status.phase}{end}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error while fetching the status phase of the function pod when verifying function is running: %v", string(stdoutStderr))
			}
			functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName, "-ojsonpath={.items[0].metadata.name}")
			functionPodName, err := functionPodsCmd.CombinedOutput()
			if err != nil {
				log.Printf("Error in fetching function pod when verifying function is running: %v", string(functionPodName))
			}
			hpaOutput, err := checkFunctionHpa()
			if err != nil {
				log.Printf("Error in fetching function hpa when verifying function is running: %v", err)
			}
			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				log.Printf("Pod: %v: is running!", string(functionPodName))
				log.Printf("Hpa: %v: is running! \n", string(hpaOutput))
				return
			}
		}
	}
}

func ensureOutputIsCorrect() {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			log.Fatalf("Timeout: test hpa failed!")
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
					functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName, "-ojsonpath={.items[0].metadata.name}")
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
					log.Printf("Tick: Response code is: %v", resp.StatusCode)
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Printf("Tick: Unable to get response: %v", err)
					}
					log.Printf("Tick: Response body is: %v", string(bodyBytes))
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

func printResponse(respCh chan string, doneCh chan bool) {
	for resp := range respCh {
		log.Println(resp)
	}
	doneCh <- true
	log.Println("All calls done!")
}

const lettersAndNums = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = lettersAndNums[rand.Intn(len(lettersAndNums))]
	}
	return string(b)
}

func makeHttpRequest(respCh chan<- string) {
	start := time.Now()
	// GET request
	testID := randomString(8)
	resp, err := http.Post(endpoint, "text/plain", bytes.NewBuffer([]byte(testID)))
	secs := time.Since(start).Seconds()
	if err != nil {
		testResult.errorRequest = fmt.Sprintf("%.2f elapsed with error on response [ERROR] %v", secs, err)
		respCh <- testResult.errorRequest
		testResult.numFailedRequests++
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		testResult.errorResponse = fmt.Sprintf("%.2f elapsed with error Unable to get response [ERROR] %v", secs, err)
		respCh <- testResult.errorResponse
		testResult.numFailedRequests++
		return
	}
	if resp.StatusCode != http.StatusOK {
		testResult.errorResponse = fmt.Sprintf("%.2f elapsed with not 200 response. response code: %s endpoint: %s", secs, resp.StatusCode, endpoint)
		respCh <- testResult.errorResponse
		testResult.numFailedRequests++
		return
	}
	respCh <- fmt.Sprintf("%.2f elapsed with response: %s response code: %s endpoint: %s", secs, string([]byte(body)), resp.StatusCode, endpoint)
	testResult.numSuccessfulRequests++
	testResult.totalRequests++
}

func getHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func cleanup() {
	log.Println("Cleaning up")
	deleteFun()
	deleteNamespace()
}

func deleteFun() {
	stdoutStderr, err := deleteK8s(functionYaml)
	output := string(stdoutStderr)
	if err != nil {
		log.Fatal("Unable to delete function ", functionName, ":\n", output)
	}
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			log.Fatal("Timed out waiting for ", functionName, " pod to be deleted\n")
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+functionName)
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(stdoutStderr), "No resources found") {
				return
			}
		}
	}
}

func deleteNamespace() {
	stdoutStderr, err := deleteK8s(nsYaml)
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "not found") {
		log.Fatal("Unable to delete namespace ", namespace, ":\n", output)
	}
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		cmd := exec.Command("kubectl", "get", "ns", namespace, "-oyaml")
		select {
		case <-timeout:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to get ns: %v\n", string(stdoutStderr))
			}
			log.Printf("Current state of the ns: %s is:\n %v", namespace, string(stdoutStderr))
			log.Fatal("Timed out waiting for namespace: ", namespace, " to be deleted\n")
		case <-tick:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil && strings.Contains(string(stdoutStderr), "NotFound") {
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

func checkFunctionAutoscaled() {
	functionHpaCmd := exec.Command("kubectl", "-n", namespace, "get", "hpa", "-l", "function="+functionName, "-ojsonpath={.items[0].metadata.name} {.items[0].spec.minReplicas} {.items[0].status.currentReplicas} {.items[0].status.currentCPUUtilizationPercentage}")
	hpaOutput, err := functionHpaCmd.CombinedOutput()
	if err != nil {
		testResult.resultMessage = fmt.Sprintf("Error in fetching function hpa: %v \n", err)
		log.Printf(testResult.resultMessage)
	} else {
		result := "Function autoscale failed"
		status := strings.Split(strings.TrimSpace(string(hpaOutput)), " ")
		minReplicas, err := strconv.Atoi(status[1])
		if err != nil {
			minReplicas = 0
		}
		minReplicasStatus := fmt.Sprintf("MinReplicas: %v", minReplicas)
		currentReplicas, err := strconv.Atoi(status[2])
		if err != nil {
			currentReplicas = 0
		}
		currentReplicasStatus := fmt.Sprintf("CurrentReplicas: %v", currentReplicas)

		cpuStatus := fmt.Sprintf("CurrentCPUUtilizationPercentage: %v%s", 0, "%")
		if len(status) == 4 {
			currentCPUUtilizationPercentage, err := strconv.Atoi(status[3])
			if err != nil {
				currentCPUUtilizationPercentage = 0
			}
			cpuStatus = fmt.Sprintf("CurrentCPUUtilizationPercentage: %v%s", currentCPUUtilizationPercentage, "%")
		}
		if currentReplicas > minReplicas {
			result = "Function autoscale succeeded"
		}
		finalStatus := fmt.Sprintf("Test HPA final status: %s \n%s \n%s \n%s\n", result, minReplicasStatus, currentReplicasStatus, cpuStatus)
		testResult.resultMessage = finalStatus
		if testResult.totalRequests > 0 {
			totalRequests := fmt.Sprintf("Requests: %v \n", testResult.totalRequests)
			testResult.resultMessage = fmt.Sprintf("%s %s", testResult.resultMessage, strings.TrimSpace(totalRequests))
		}
		if testResult.numFailedRequests > 0 {
			numFailedRequests := fmt.Sprintf("Failed resquests: %v \n", testResult.numFailedRequests)
			testResult.resultMessage = fmt.Sprintf("%s %s", testResult.resultMessage, strings.TrimSpace(numFailedRequests))
		}
		if testResult.numSuccessfulRequests > 0 {
			numSuccessfulRequests := fmt.Sprintf("Successful requests: %v \n", testResult.numSuccessfulRequests)
			testResult.resultMessage = fmt.Sprintf("%s %s", testResult.resultMessage, strings.TrimSpace(numSuccessfulRequests))
		}
		if len(testResult.errorRequest) > 0 {
			errorRequest := fmt.Sprintf("Request error: %v \n", testResult.errorRequest)
			testResult.resultMessage = fmt.Sprintf("%s %s", testResult.resultMessage, strings.TrimSpace(errorRequest))
		}
		if len(testResult.errorResponse) > 0 {
			errorResponse := fmt.Sprintf("Response error: %v \n", testResult.errorResponse)
			testResult.resultMessage = fmt.Sprintf("%s %s", testResult.resultMessage, strings.TrimSpace(errorResponse))
		}
		log.Println(testResult.resultMessage)
	}
}

func (slack *Slack) sendNotificationtoSlackChannel(testResult *TestResult) {
	textMessage := fmt.Sprintf(`{"channel": "%v", "text":"%v"}"`, slack.SlackChaneel, testResult.resultMessage)
	var jsonStr = []byte(textMessage)
	req, err := http.NewRequest("POST", slack.SlackEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to send Slack notification to endpoint: %v : Error: %v", slack.SlackChaneel, err)
	}
	defer resp.Body.Close()
	log.Println("Slack response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("Slack response response Body:", string(body))
}

func NewSlack() *Slack {
	sUrl := os.Getenv("LOAD_TEST_SLACK_ENDPOINT")
	if len(sUrl) == 0 {
		sUrl = slackEndpoint
	}
	sChannel := os.Getenv("LOAD_TEST_SLACK_CHANNEL")
	if len(sChannel) == 0 {
		sChannel = slackChaneel
	}
	s := &Slack{sUrl, sChannel}

	log.Printf("Slack: %v", s )
	return s
}

func NewTestResult() *TestResult {
	t := &TestResult{}
	return t
}
