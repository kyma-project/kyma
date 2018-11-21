package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func deployK8s(yamlFile string) {
	cmd := exec.Command("kubectl", "create", "-f", yamlFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy:\n", string(stdoutStderr))
	}
}

func deleteK8s(yamlFile string) {
	cmd := exec.Command("kubectl", "delete", "-f", yamlFile, "--grace-period=0", "--force", "--ignore-not-found")
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		log.Fatal("Unable to delete:\n", output)
	}
}

func printContentsOfNamespace(namespace string) {
	getResourcesCmd := exec.Command("kubectl", "-n", namespace, "get", "all,function")
	stdoutStderr, err := getResourcesCmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		log.Fatal("Unable to get all,function:\n", output)
	}
	log.Printf("Current contents of the ns:%s is:\n %s", namespace, output)
}

func deleteNamespace(namespace string) {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)

	deleteCmd := exec.Command("kubectl", "delete", "ns", namespace, "--grace-period=0", "--force", "--ignore-not-found")
	stdoutStderr, err := deleteCmd.CombinedOutput()

	if err != nil && !strings.Contains(string(stdoutStderr), "No resources found") && !strings.Contains(string(stdoutStderr), "Error from server (Conflict): Operation cannot be fulfilled on namespaces") {
		log.Fatalf("Error while deleting namespace: %s, to be deleted\n Output:\n%s", namespace, string(stdoutStderr))
	}

	for {
		cmd := exec.Command("kubectl", "get", "ns", namespace, "-oyaml")
		select {
		case <-timeout:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to get ns: %v\n", string(stdoutStderr))
			}
			log.Printf("Current state of the ns: %s is:\n %v", namespace, string(stdoutStderr))

			printContentsOfNamespace(namespace)
			log.Fatal("Timed out waiting for namespace: ", namespace, " to be deleted\n")
		case <-tick:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil && strings.Contains(string(stdoutStderr), "NotFound") {
				return
			}
		}
	}
}

func deployFun(namespace, name, runtime, codeFile, handler, deps string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "deploy", name, "-r", runtime, "-f", codeFile, "--handler", handler, "-d", deps)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy function ", name, ":\n", string(stdoutStderr))
	}
	ensureFunctionIsRunning(namespace, name)
}

func ensureFunctionIsRunning(namespace, name string) {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "describe", "pod", "-l", "function="+name)
			stdoutStderr, _ := cmd.CombinedOutput()
			printLogsFunctionPodContainers(namespace, name)
			log.Fatalf("Timed out waiting for: %v function pod to be running. Because of following error: %v ", name, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={range .items[*]}{.status.phase}{end}")

			stdoutStderr, err := cmd.CombinedOutput()

			if err != nil {
				log.Printf("Error while fetching the status phase of the function pod when verifying function is running: %v", string(stdoutStderr))
			}

			functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")

			functionPodName, err := functionPodsCmd.CombinedOutput()
			if err != nil {
				log.Printf("Error is fetch function pod when verifying function is running: %v", string(functionPodName))
			}
			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				log.Printf("Pod: %v: is running!", string(functionPodName))
				return
			}
		}
	}
}

func printLogsFunctionPodContainers(namespace, name string) {
	functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")
	functionPodName, err := functionPodsCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error is fetch function pod: %v", string(functionPodName))
	}

	log.Printf("---------- Logs from all containers for function pod: %s ----------\n", string(functionPodName))

	prepareContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), "prepare")

	prepareContainerLog, _ := prepareContainerLogCmd.CombinedOutput()
	log.Printf("Logs from prepare container:\n%s\n", string(prepareContainerLog))

	installContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), "install")

	installContainerLog, _ := installContainerLogCmd.CombinedOutput()
	log.Printf("Logs from prepare container:\n%s\n", string(installContainerLog))

	functionContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), name)

	functionContainerLog, _ := functionContainerLogCmd.CombinedOutput()
	log.Printf("Logs from %s container in pod %s:\n%s\n", name, string(functionPodName), string(functionContainerLog))

	envoyLogsCmd := exec.Command("kubectl", "-n", namespace, "log", "-l", string(functionPodName), "-c", "istio-proxy")

	envoyLogsCmdStdErr, _ := envoyLogsCmd.CombinedOutput()
	log.Printf("Envoy Logs are:\n%s\n", string(envoyLogsCmdStdErr))
}

func deleteFun(namespace, name string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "delete", name)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "not found") {
		log.Fatal("Unable to delete function ", name, ":\n", output)
	}

	cmd = exec.Command("kubectl", "-n", namespace, "delete", "pod", "-l", "function="+name, "--grace-period=0", "--force")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(stdoutStderr), "No resources found") && !strings.Contains(string(stdoutStderr), "warning: Immediate deletion does not wait for confirmation that the running resource has been terminated") {
		log.Fatal("Unable to delete function pod:\n", string(stdoutStderr))
	}

	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			log.Fatal("Timed out waiting for ", name, " pod to be deleted\n")
		case <-tick:
			cmd = exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name)
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(stdoutStderr), "No resources found") {
				return
			}
		}
	}
}

func ensureOutputIsCorrect(host, expectedOutput, testID, namespace, name string) {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			printLogsFunctionPodContainers(namespace, name)
			log.Fatalf("Timeout: kubeless test failed!")

		case <-tick:
			resp, err := http.Post(host, "text/plain", bytes.NewBuffer([]byte(testID)))
			if err != nil {
				log.Printf("Unable to call host: %v : Error: %v", host, err)
			} else {
				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatalf("Unable to get response: %v", err)
					}
					log.Printf("Response from function: %v\n", string(bodyBytes))

					functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")
					functionPodName, err := functionPodsCmd.CombinedOutput()
					if err != nil {
						log.Printf("Error in fetch function pod when verifying correct output: %v", string(functionPodName))
					}
					if strings.EqualFold(expectedOutput, string(bodyBytes)) {
						log.Printf("Response is equal to expected output: %v == %v", string(bodyBytes), expectedOutput)
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

const lettersAndNums = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = lettersAndNums[rand.Intn(len(lettersAndNums))]
	}
	return string(b)
}

func cleanup(namespace, functionName string) {
	log.Println("Cleaning up")
	deleteFun(namespace, functionName)
	deleteNamespace(namespace)
}

func main() {
	const namespace = "kubeless-unit"
	const functionName = "test-hello"
	testID := randomString(8)

	cleanup(namespace, functionName)

	rand.Seed(time.Now().UTC().UnixNano())

	log.Println("Starting kubeless unit test")

	log.Printf("Creating namespace: %v\n", namespace)
	deployK8s("ns.yaml")

	log.Println("Deploying test-hello function")
	deployFun(namespace, functionName, "nodejs8", "hello.js", "hello.main", "package.json")

	log.Println("Verifying correct function output for test-hello")
	host := fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
	log.Printf("Endpoint for the function: %v\n", host)
	expectedOutput := fmt.Sprintf("{\"result\":\"%v\"}", testID)
	ensureOutputIsCorrect(host, expectedOutput, testID, namespace, functionName)

	cleanup(namespace, functionName)
	log.Println("Kubeless unit test is successful")
}
