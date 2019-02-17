package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	slackApi "github.com/nlopes/slack"
)

const (
	functionName   = "load-test"
	namespace      = "load-test"
	functionYaml   = "k8syaml/function.yaml"
	nsYaml         = "k8syaml/ns.yaml"
	expectedOutput = "Call to the function load-test was successful!"
)

type testResult struct {
	sync.RWMutex
	resultMessage         string
	errorResponse         string
	errorRequest          string
	numFailedRequests     int
	numSuccessfulRequests int
	totalRequestsDone     int
	totalRequests         int
}

var (
	numRequest    int
	endpoint      = fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
	client        = getHTTPClient(true)
	reporter      *Reporter
	tResult       *testResult
	timeout       = time.After(time.Duration(5) * time.Minute)
	stopping      = false
	mutex         sync.RWMutex
	slackUserName = "load-test"
	slackEmoji    = ":weight_lifter"
)

func init() {
	cleanup()
	log.Printf("Creating namespace: %s", namespace)
	createNS()
	log.Printf("Deploying function: %s", functionName)
	deployFun()
	log.Printf("Verifying correct function output for %s", functionName)
	log.Printf("HTTP endpoint for the function: %v", endpoint)
	ensureOutputIsCorrect()

	numRequestEnv := os.Getenv("LOAD_TEST_TOTAL_REQS_PER_ROUTINE")
	if len(numRequestEnv) == 0 {
		numRequest = 1600
	} else {
		num, err := strconv.Atoi(numRequestEnv)
		if err != nil {
			log.Fatalf("Failed to convert value of LOAD_TEST_TOTAL_REQS_PER_ROUTINE to int: %v", err)
		}
		numRequest = num
	}

	token := os.Getenv("LOAD_TEST_SLACK_TOKEN")
	if len(token) == 0 {
		log.Fatalln("No slack api token provided!")
	}
	channel := os.Getenv("LOAD_TEST_SLACK_CHANNEL")
	if len(channel) == 0 {
		log.Fatalln("No slack channel provided!")
	}
	slackCli := slackApi.New(token)

	reporter = NewReporter(*slackCli, channel, slackUserName, slackEmoji)
	tResult = newTestResult()
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
	tResult.totalRequests = numCPUs * numRequest
	log.Println("Number of requests to be done: ", tResult.totalRequests)
	for c := 0; c < numCPUs; c++ {
		go func() {
			defer wg.Done()
			for r := 0; r < numRequest; r++ {
				mutex.RLock()
				if stopping {
					break
				}
				mutex.RUnlock()
				makeHTTPRequest(respCh)
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

func makeHTTPRequest(respCh chan<- string) {
	tResult.Lock()
	defer tResult.Unlock()
	tResult.totalRequestsDone++
	start := time.Now()
	testID := randomString(8)
	resp, err := http.Post(endpoint, "text/plain", bytes.NewBuffer([]byte(testID)))
	secs := time.Since(start).Seconds()
	if err != nil {
		tResult.errorRequest = fmt.Sprintf("TestId: [%v] %.2f secs elapsed with error on response [ERROR] %v", testID, secs, err)
		respCh <- tResult.errorRequest
		tResult.numFailedRequests++
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tResult.errorResponse = fmt.Sprintf("TestId: [%v] %.2f secs elapsed with error : unable to read response [ERROR] %v", testID, secs, err)
		respCh <- tResult.errorResponse
		tResult.numFailedRequests++
		return
	}
	if resp.StatusCode != http.StatusOK {
		tResult.errorResponse = fmt.Sprintf("TestId: [%v] %.2f secs elapsed with no 200 response: response code: %v endpoint: %s", testID, secs, resp.StatusCode, endpoint)
		respCh <- tResult.errorResponse
		tResult.numFailedRequests++
		return
	}
	respCh <- fmt.Sprintf("TestId: [%v] Response code: HTTP %v, Response: %s, Response time: %.2f secs,  endpoint: %s", testID, resp.StatusCode, string([]byte(body)), secs, endpoint)
	tResult.numSuccessfulRequests++
}

func closingTest(start time.Time, premature bool) {
	checkFunctionAutoscaled(premature)
	reporter.Report(tResult.resultMessage, premature)
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

func newTestResult() *testResult {
	t := &testResult{}
	return t
}
