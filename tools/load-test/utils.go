package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

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

func getHTTPClient(skipVerify bool) *http.Client {
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

	timeoutDelFunc := time.After(15 * time.Minute)
	tick := time.Tick(1 * time.Second)
	done := false
	for {
		var err error
		select {
		case <-timeoutDelFunc:
			log.Fatalf("Unable to delete function: %s error: %v \n", functionName, err)
		case <-tick:
			_, err = deleteK8s(functionYaml)
			if err == nil {
				log.Println("Execution of delete function is successful!")
				done = true
			}
		}
		if done {
			break
		}
	}

	for {
		select {
		case <-timeoutDelFunc:
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
				log.Fatalf("Unable to get resources in ns: %v\n", string(stdoutStderr))
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
	tResult.RLock()
	functionHpaCmd := exec.Command("kubectl", "-n", namespace, "get", "hpa", "-l", "function="+functionName, "-ojsonpath={.items[0].metadata.name} {.items[0].spec.minReplicas} {.items[0].status.currentReplicas} {.items[0].status.currentCPUUtilizationPercentage}")
	hpaOutput, err := functionHpaCmd.CombinedOutput()
	result := ""
	if err != nil {
		tResult.resultMessage = fmt.Sprintf("Error in fetching function HPA: %v \n", string(hpaOutput))
		log.Printf(tResult.resultMessage)
	} else {
		if premature {
			result = "<!channel> Horizontal Pod Autoscaler test timed out!"
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
		finalStatus := fmt.Sprintf("Status: %s \n%s \n%s \n%s\n", result, minReplicasStatus, currentReplicasStatus, cpuStatus)
		tResult.resultMessage = finalStatus

		if tResult.totalRequests > 0 {
			totalRequests := fmt.Sprintf("Total number of requests: %v \n", tResult.totalRequests)
			tResult.resultMessage = fmt.Sprintf("%s %s\n", tResult.resultMessage, strings.TrimSpace(totalRequests))
		}

		if tResult.totalRequestsDone > 0 {
			totalRequestsDone := fmt.Sprintf("Total number of requests done: %v \n", tResult.totalRequestsDone)
			tResult.resultMessage = fmt.Sprintf("%s %s\n", tResult.resultMessage, strings.TrimSpace(totalRequestsDone))
		}

		if tResult.numSuccessfulRequests > 0 {
			numSuccessfulRequests := fmt.Sprintf("Successful requests: %v \n", tResult.numSuccessfulRequests)
			tResult.resultMessage = fmt.Sprintf("%s %s\n", tResult.resultMessage, strings.TrimSpace(numSuccessfulRequests))
		}

		if tResult.numFailedRequests > 0 {
			numFailedRequests := fmt.Sprintf("Failed resquests: %v \n", tResult.numFailedRequests)
			tResult.resultMessage = fmt.Sprintf("%s %s\n", tResult.resultMessage, strings.TrimSpace(numFailedRequests))
			tResult.totalRequests = tResult.totalRequests + tResult.numFailedRequests
		}

		log.Println(tResult.resultMessage)
		tResult.RUnlock()
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

func createNS() {
	stdoutStderr, err := deployK8s(nsYaml)
	if err != nil {
		log.Fatal("unable to create namespace ", namespace, ":\n", string(stdoutStderr))
	}
}
