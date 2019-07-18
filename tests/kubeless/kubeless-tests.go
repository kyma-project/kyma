package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/avast/retry-go"
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
	log.Infof("Current contents of the ns:%s is:\n %s", namespace, output)
}

func deleteNamespace(namespace string) error {
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)

	deleteCmd := exec.Command("kubectl", "delete", "ns", namespace, "--grace-period=0", "--force", "--ignore-not-found")
	stdoutStderr, err := deleteCmd.CombinedOutput()

	if err != nil && !strings.Contains(string(stdoutStderr), "No resources found") && !strings.Contains(string(stdoutStderr), "Error from server (Conflict): Operation cannot be fulfilled on namespaces") {
		log.Errorf("Error while deleting namespace: %s, to be deleted\n Output:\n%s", namespace, string(stdoutStderr))
		return err
	}

	for {
		cmd := exec.Command("kubectl", "get", "ns", namespace, "-oyaml")
		select {
		case <-timeout:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("Unable to get ns: %v\n", string(stdoutStderr))
				return err
			}
			log.Infof("Current state of the ns: %s is:\n %v", namespace, string(stdoutStderr))
			printContentsOfNamespace(namespace)
			log.Error("Timed out waiting for namespace: ", namespace, " to be deleted\n")
			return fmt.Errorf("timed out waiting for namespace deletion")
		case <-tick:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil && strings.Contains(string(stdoutStderr), "NotFound") {
				return nil
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
				log.Infof("Error while fetching the status phase of the function pod when verifying function is running: %v", string(stdoutStderr))
			}

			functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")

			functionPodName, err := functionPodsCmd.CombinedOutput()
			if err != nil {
				log.Infof("Error is fetch function pod when verifying function is running: %v", string(functionPodName))
			}
			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				log.Infof("Pod: %v: is running!", string(functionPodName))
				return
			}
		}
	}
}

func printLogsFunctionPodContainers(namespace, name string) {
	functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")
	functionPodName, err := functionPodsCmd.CombinedOutput()
	if err != nil {
		log.Infof("Error is fetch function pod: %v", string(functionPodName))
	}

	log.Infof("---------- Logs from all containers for function pod: %s ----------\n", string(functionPodName))

	prepareContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), "prepare")

	prepareContainerLog, _ := prepareContainerLogCmd.CombinedOutput()
	log.Infof("Logs from prepare container:\n%s\n", string(prepareContainerLog))

	installContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), "install")

	installContainerLog, _ := installContainerLogCmd.CombinedOutput()
	log.Infof("Logs from prepare container:\n%s\n", string(installContainerLog))

	functionContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), name)

	functionContainerLog, _ := functionContainerLogCmd.CombinedOutput()
	log.Infof("Logs from %s container in pod %s:\n%s\n", name, string(functionPodName), string(functionContainerLog))

	envoyLogsCmd := exec.Command("kubectl", "-n", namespace, "log", "-l", string(functionPodName), "-c", "istio-proxy")

	envoyLogsCmdStdErr, _ := envoyLogsCmd.CombinedOutput()
	log.Infof("Envoy Logs are:\n%s\n", string(envoyLogsCmdStdErr))
}

func deleteFun(namespace, name string) error {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "delete", name)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "not found") {
		log.Error("Unable to delete function ", name, ":\n", output)
		return err
	}

	cmd = exec.Command("kubectl", "-n", namespace, "delete", "pod", "-l", "function="+name, "--grace-period=0", "--force")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(stdoutStderr), "No resources found") && !strings.Contains(string(stdoutStderr), "warning: Immediate deletion does not wait for confirmation that the running resource has been terminated") {
		log.Error("Unable to delete function pod:\n", string(stdoutStderr))
		return err
	}

	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			log.Error("Timed out waiting for ", name, " pod to be deleted\n")
			return fmt.Errorf("Pod deletion timed out")
		case <-tick:
			cmd = exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name)
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(stdoutStderr), "No resources found") {
				return nil
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
			log.Fatal("Timeout: kubeless test failed!")

		case <-tick:
			resp, err := http.Post(host, "text/plain", bytes.NewBuffer([]byte(testID)))
			if err != nil {
				log.Infof("Function not yet ready: Unable to call host: %v : Error: %v", host, err)
			} else {
				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatalf("Unable to get response: %v", err)
					}
					log.Infof("Response from function: %v\n", string(bodyBytes))

					functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")
					functionPodName, err := functionPodsCmd.CombinedOutput()
					if err != nil {
						log.Errorf("Error in fetch function pod when verifying correct output: %v", string(functionPodName))
					}
					if strings.EqualFold(expectedOutput, string(bodyBytes)) {
						log.Infof("Response is equal to expected output: %v == %v", string(bodyBytes), expectedOutput)
						log.Infof("Name of the successful pod is: %v", string(functionPodName))
						return
					}
					log.Infof("Name of the failed pod is: %v", string(functionPodName))
					log.Fatalf("Response is not equal to expected output:\nResponse: %v\nExpected: %v", string(bodyBytes), expectedOutput)
				} else {
					log.Infof("Tick: Response code is: %v", resp.StatusCode)
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Infof("Tick: Unable to get response: %v", err)
					}
					log.Infof("Tick: Response body is: %v", string(bodyBytes))
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

func cleanup(namespace, functionName string) error{
	log.Info("Cleaning up")
	err:=deleteFun(namespace, functionName)
	if err != nil {
		return err
	}
	err = deleteNamespace(namespace)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetReportCaller(true)

	const namespace = "kubeless-unit"
	const functionName = "test-hello"
	testID := randomString(8)

	log.Info("Pre Test Cleanup")
	err:=retry.Do(func() error {
		return cleanup(namespace, functionName)
	}, retry.Attempts(5), retry.Delay(5 * time.Second))
	if err != nil {
		log.WithField("error", err).Fatal("Cleanup failed")
	}

	rand.Seed(time.Now().UTC().UnixNano())

	log.Info("Starting kubeless unit test")

	log.Infof("Creating namespace: %v\n", namespace)
	deployK8s("ns.yaml")

	log.Info("Deploying test-hello function")
	deployFun(namespace, functionName, "nodejs8", "hello.js", "hello.main", "package.json")

	log.Info("Verifying correct function output for test-hello")
	host := fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
	log.Infof("Endpoint for the function: %v\n", host)
	expectedOutput := fmt.Sprintf("{\"result\":\"%v\"}", testID)
	ensureOutputIsCorrect(host, expectedOutput, testID, namespace, functionName)

	log.Info("Post Test Cleanup")
	cleanup(namespace, functionName)
	log.Info("Kubeless unit test is successful")
}
