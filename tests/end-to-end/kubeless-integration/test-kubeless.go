package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/common/ingressgateway"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
	getResourcesCmd := exec.Command("kubectl", "-n", namespace, "get", "deployments,services,replicasets,pods,serviceinstances,servicebindings,servicebindingusages,functions,subscriptions.eventing.kyma-project.io,apis.gateway.kyma-project.io,eventactivations.applicationconnector.kyma-project.io")
	stdoutStderr, err := getResourcesCmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		log.Fatal("Unable to get deployments,services,replicasets,pods,serviceinstances,servicebindings,servicebindingusages,functions,subscriptions.eventing.kyma-project.io,apis.gateway.kyma-project.io,eventactivations.applicationconnector.kyma-project.io:\n", output)
	}
	log.Printf("Current contents of the ns:%s is:\n %v", namespace, output)
}

func deleteNamespace(namespace string) {
	timeout := time.After(6 * time.Minute)
	tick := time.Tick(1 * time.Second)

	deleteCmd := exec.Command("kubectl", "delete", "ns", namespace, "--grace-period=0", "--force", "--ignore-not-found")
	stdoutStderr, err := deleteCmd.CombinedOutput()

	if err != nil && !strings.Contains(string(stdoutStderr), "No resources found") && !strings.Contains(string(stdoutStderr), "Error from server (Conflict): Operation cannot be fulfilled on namespaces") {
		log.Fatalf("Error while deleting namespace: %s, to be deleted\n Output:\n%s", namespace, string(stdoutStderr))
	}

	log.Printf("Current state of the ns:%s is:\n %v", namespace, string(stdoutStderr))

	for {
		cmd := exec.Command("kubectl", "get", "ns", namespace, "-oyaml")
		select {
		case <-timeout:
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatal("Unable to get ns:\n", string(stdoutStderr))
			}
			log.Printf("Current state of the ns:%s is:\n %v", namespace, string(stdoutStderr))

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

func deployFun(namespace, name, runtime, codeFile, handler string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "deploy", name, "-r", runtime, "-f", codeFile, "--handler", handler, "--memory", "128Mi")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy function ", name, ":\n", string(stdoutStderr))
	}
	ensureFunctionIsRunning(namespace, name, false)
}

var uidRegex = regexp.MustCompile(`^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)

func getSBUID(namespace, name string) string {
	timeout := time.After(1 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			log.Fatal("Timed out getting servicebindingusage ", name, "\n")
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "servicebindingusage", name, "-ojsonpath={.metadata.uid}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil {
				submatches := uidRegex.FindStringSubmatch(string(stdoutStderr))
				if submatches != nil {
					return submatches[1]
				}
			}
		}
	}
}

func ensureFunctionIsRunning(namespace, name string, serviceBinding bool) {
	sbuID := ""
	if serviceBinding {
		sbuID = getSBUID(namespace, name)
		log.Printf("[%v] Service binding Usage ID is: %v", name, sbuID)
	}
	timeout := time.After(6 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "describe", "pod", "-l", "function="+name)
			if sbuID != "" {
				log.Printf("[%v] Timed out: Service binding Usage ID is: %v", name, sbuID)
				cmd = exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-l", "use-"+sbuID)
				printDebugLogsSvcBindingUsageFailed(namespace, name)
			}
			stdoutStderr, _ := cmd.CombinedOutput()
			printLogsFunctionPodContainers(namespace, name)
			log.Fatalf("[%v] Timed out waiting for: %v function pod to be running. Because of following error: %v ", name, name, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={range .items[*]}{.status.phase}{end}")
			if sbuID != "" {
				log.Printf("[%v] Tick: Service binding Usage ID is: %v", name, sbuID)
				cmd = exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-l", "use-"+sbuID, "-ojsonpath={range .items[*]}{.status.phase}{end}")
			}
			stdoutStderr, err := cmd.CombinedOutput()

			if err != nil {
				log.Printf("[%v] Error while fetching the status phase of the function pod when verifying function is running: %v", name, string(stdoutStderr))
			}

			functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")

			functionPodName, err := functionPodsCmd.CombinedOutput()
			if err != nil {
				log.Printf("[%v] Error is fetch function pod when verifying function is running: %v", name, string(functionPodName))
			}
			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				log.Printf("[%v] Pod: %v: and SBU ID is: %v", name, string(functionPodName), sbuID)
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

func printDebugLogsAPIServiceFailed(namespace, name string) {
	functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name)
	functionPodsStdOutStdErr, err := functionPodsCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching pods for function: %v\n", err)
	}
	log.Printf("Function pods status:\n%s\n", string(functionPodsStdOutStdErr))

	functionSvcCmd := exec.Command("kubectl", "-n", namespace, "get", "svc", "-l", "function="+name)
	functionSvcCmdStdOutStdErr, err := functionSvcCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching service for function: %v\n", err)
	}
	log.Printf("Function service status:\n%s\n", string(functionSvcCmdStdOutStdErr))

	apiListCmd := exec.Command("kubectl", "-n", namespace, "get", "api", "-l", "function="+name)
	apiListStdOutErr, err := apiListCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching list APIs: %v\n", err)
	}
	log.Printf("API List:\n%s\n", string(apiListStdOutErr))

	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	apiControllerPodNameCmd := exec.Command("kubectl", "-n", controllerNamespace, "get", "po", "-l", "app=api-controller", "-o", "jsonpath={.items[0].metadata.name}")

	apiControllerPodName, err := apiControllerPodNameCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching API Controller pod: \n%s\n", string(apiControllerPodName))
	}

	apiControllerLogsCmd := exec.Command("kubectl", "-n", controllerNamespace, "log", string(apiControllerPodName))

	apiControllerLogsCmdOutErr, err := apiControllerLogsCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching logs for API Controller: %v\n", string(apiControllerLogsCmdOutErr))
	}
	log.Printf("Logs from API Controller:\n%s\n", string(apiControllerLogsCmdOutErr))
}

func printDebugLogsSvcBindingUsageFailed(namespace, name string) {

	functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name)

	functionPodsStdOutStdErr, err := functionPodsCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching pods for function: %v\n", err)
	}
	log.Printf("Function pods status:\n%s\n", string(functionPodsStdOutStdErr))

	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	svcBindingUsageControllerPodNameCmd := exec.Command("kubectl", "-n", controllerNamespace, "get", "po", "-l", "app=service-binding-usage-controller", "-o", "jsonpath={.items[0].metadata.name}")

	svcBindingUsageControllerPodName, err := svcBindingUsageControllerPodNameCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching servicebindingusagercontroller pod: \n%s\n", string(svcBindingUsageControllerPodName))
	}

	svcBindingUsageControllerLogsCmd := exec.Command("kubectl", "-n", controllerNamespace, "log", string(svcBindingUsageControllerPodName), "-c", "service-binding-usage-controller")

	svcBindingUsageControllerLogsCmdOutErr, err := svcBindingUsageControllerLogsCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching logs for bindingusagecontroller: %v\n", string(svcBindingUsageControllerLogsCmdOutErr))
	}
	log.Printf("Logs from bindingusagecontroller:\n%s\n", string(svcBindingUsageControllerLogsCmdOutErr))

}

func connectUsingK8sService(namespace, name string) {
	log.Printf("[%v] Trying to curl using local kubernetes service", name)
	url := "http://" + name + "." + namespace + ":8080"
	cmd := exec.Command("curl", "-v", url)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[%v] Unable to curl to function using internal kubernetes service: %v", name, string(stdoutStderr))
		return
	}
	log.Printf("[%v] Result of curl request is: %v", name, string(stdoutStderr))
	return
}

func printDebugLogsForEvents() {
	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")

	eventsPodCmd := exec.Command("kubectl", "-n", controllerNamespace, "logs", "-l", "app=push", "-c", "push")

	eventsPodStdErr, err := eventsPodCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while fetching logs for event-bus-push: %v\n", string(eventsPodStdErr))
	}
	log.Printf("Event logs from 'Push' podare:\n%s\n", string(eventsPodStdErr))
}

func deleteFun(namespace, name string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "delete", name)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "not found") {
		log.Fatal("Unable to delete function ", name, ":\n", output)
	}

	timeout := time.After(6 * time.Minute)
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

func ensureOutputIsCorrect(host, expectedOutput, testID, namespace, testName string) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)

	ingressClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		log.Fatalf("Cannot get ingressgateway address: %s", err)
	}

	for {
		select {
		case <-timeout:
			log.Printf("[%v] Timeout: Check if Virtual Service has been created", testName)

			printDebugLogsAPIServiceFailed(namespace, testName)
			connectUsingK8sService(namespace, testName)
			printLogsFunctionPodContainers(namespace, testName)

			cmd := exec.Command("kubectl", "-n", namespace, "get", "virtualservices.networking.istio.io", "-l", "apiName="+testName)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("[%v] Unable to fetch Virtual Service for: %v. Because of following error: %v", testName, testName, string(stdoutStderr))
			}
			log.Printf("[%v] Timeout: Getting Virtual Service: '%v' and result is: %v", testName, testName, string(stdoutStderr))

			log.Printf("[%v] Timeout: Check http Get one last time", testName)
			resp, err := ingressClient.Post(host, "text/plain", bytes.NewBuffer([]byte(testID)))
			if err != nil {
				log.Fatalf("[%v] Timeout: Unable to call host: %v. Because of following error: %v ", testName, host, err)
			}
			log.Fatalf("[%v] Timeout: Response is: %v", testName, resp)

		case <-tick:
			resp, err := ingressClient.Post(host, "text/plain", bytes.NewBuffer([]byte(testID)))
			if err != nil {
				log.Fatalf("[%v] Unable to call host: %v. Because of following error: %v", testName, host, err)
			}
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatalf("[%v] Unable to get response: %v", testName, err)
				}

				functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+testName, "-ojsonpath={.items[0].metadata.name}")
				functionPodName, err := functionPodsCmd.CombinedOutput()
				if err != nil {
					log.Printf("[%v] Error is fetch function pod when verifying correct output: %v", testName, string(functionPodName))
				}
				if string(bodyBytes) == expectedOutput {
					log.Printf("[%v] Response is equal to expected output: %v == %v", testName, string(bodyBytes), expectedOutput)
					log.Printf("[%v] Name of the Successful Pod is: %v", testName, string(functionPodName))
					return
				}
				log.Printf("[%v] Response is not equal to expected output: %v != %v, pod name: %s. Retry...", testName, string(bodyBytes), expectedOutput, string(functionPodName))
			} else {
				log.Printf("[%v] Tick: Response code is: %v", testName, resp.StatusCode)
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("[%v] Tick: Unable to get response: %v", testName, err)
				}
				log.Printf("[%v] Tick: Response body is: %v", testName, string(bodyBytes))
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

func publishEvent(testID string) {
	cmd := exec.Command("curl", "-s", "http://event-publish-service.kyma-system:8080/v1/events", "-H", "Content-Type: application/json", "-d", `{"source-id": "dummy", "event-type": "test", "event-type-version": "v1", "event-time": "0001-01-01T00:00:00+00:00", "data": "`+testID+`"}`)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to publish event(error: %s): %s\n", err, string(stdoutStderr))
	}
}

func ensureCorrectLog(namespace, funName string, pattern *regexp.Regexp, match string, serviceBinding bool) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)
	sbuID := ""
	if serviceBinding {
		sbuID = getSBUID(namespace, funName)
	}
	for {
		var cmd *exec.Cmd
		if sbuID == "" {
			cmd = exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", "function="+funName, "--field-selector=status.phase==Running", "--output=jsonpath={.items[0].metadata.name}")
		} else {
			cmd = exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", "use-"+sbuID+","+"function="+funName, "--field-selector=status.phase==Running", "--output=jsonpath={.items[0].metadata.name}")
		}

		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Unable to get running pod for function: %v due to following reason: %v", funName, err)
		}
		pod := string(stdoutStderr)
		log.Printf("Checking logs for pods: %v", pod)
		select {
		case <-timeout:
			log.Printf("[%s] Timeout printing debug logs", funName)
			if !serviceBinding {
				printDebugLogsForEvents()
			}
			cmd := exec.Command("kubectl", "-n", namespace, "logs", pod, "-c", funName)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatal("Timed out getting the correct log for ", funName, ":\n", string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "logs", pod, "-c", funName)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatal("Unable to obtain function log for ", funName, ":\n", string(stdoutStderr))
			}

			submatches := pattern.FindStringSubmatch(string(stdoutStderr))
			if submatches != nil && submatches[1] == match {
				return
			}
		}
	}
}

func ensureSvcInstanceIsDeployed(namespace, svcInstance string) {
	timeout := time.After(6 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "describe", "-n", namespace, "serviceinstance", svcInstance)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to fetch service instance %v:\n%v", svcInstance, string(stdoutStderr))
			}
			log.Fatalf("Timeout waiting to get service instance ProvisionedSuccessfully %v: %v", svcInstance, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "serviceinstance", svcInstance, "-o=jsonpath={.items[*]}{.status.conditions[*].reason}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error fetching service instance %v: %v", svcInstance, string(stdoutStderr))
			}
			if string(stdoutStderr) == "ProvisionedSuccessfully" {
				return
			}
		}
	}
}

func ensureServceBindingIsReady(namespace, svcBinding string) {
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "describe", "-n", namespace, "servicebinding", svcBinding)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to fetch service instance binding %v:\n%v", svcBinding, string(stdoutStderr))
			}
			log.Fatalf("Timeout waiting to get service instance binding ProvisionedSuccessfully %v: %v", svcBinding, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "servicebinding", svcBinding, "-o=jsonpath={.items[*]}{.status.conditions[*].status}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error fetching service instance binding %v: %v", svcBinding, err)
			}
			if string(stdoutStderr) == "True" {
				log.Printf("Service binding has been successfully created.")
				return
			}
		}
	}
}

func cleanup(cl client.Client) {
	log.Println("Cleaning up")
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		deleteK8s("k8syaml/k8s.yaml")
	}()
	go func() {
		defer wg.Done()
		deleteFun("kubeless-test", "test-hello")
	}()
	go func() {
		defer wg.Done()
		deleteFun("kubeless-test", "test-event")
	}()
	go func() {
		defer wg.Done()
		deleteK8s("svc-binding.yaml")
		deleteK8s("k8syaml/svcbind-lambda.yaml")
		deleteK8s("svc-instance.yaml")
		deleteAddonsConfiguration(cl, "kubeless-test", "kubeless-addon-redis")
	}()
	wg.Wait()
	deleteNamespace("kubeless-test")
}

var testDataRegex = regexp.MustCompile(`(?m)^OK ([a-z0-9]{8})$`)

func main() {
	cl, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		log.Fatalf("Error getting kuberentes client config: %v", err)
	}

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Error registering addons configuration scheme: %v", err)
	}

	cleanup(cl)

	time.Sleep(10 * time.Second)
	rand.Seed(time.Now().UTC().UnixNano())

	log.Println("Starting test")
	log.Printf("Domain Name is: %v", os.Getenv("DOMAIN_NAME"))

	testID := randomString(8)
	deployK8s("ns.yaml")
	deployK8s("k8syaml/k8s.yaml")
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		log.Println("Deploying test-hello function")
		deployFun("kubeless-test", "test-hello", "nodejs6", "hello.js", "hello.handler")
		log.Println("Verifying correct function output for test-hello")
		host := fmt.Sprintf("https://test-hello.%s", os.Getenv("DOMAIN_NAME"))
		ensureOutputIsCorrect(host, "hello world", testID, "kubeless-test", "test-hello")
		log.Println("Function test-hello works correctly")
	}()

	go func() {
		defer wg.Done()
		log.Println("Deploying test-event function")
		deployFun("kubeless-test", "test-event", "nodejs6", "event.js", "event.handler")
		time.Sleep(2 * time.Minute) // Sometimes subsctiptions take long time. So lambda might not get the events
		log.Println("Publishing event to function test-event")
		publishEvent(testID)
		log.Println("Verifying correct event processing for test-event")
		ensureCorrectLog("kubeless-test", "test-event", testDataRegex, testID, false)
		log.Println("Function test-event works correctly")
	}()

	go func() {
		defer wg.Done()
		log.Println("Deploying AddonsConfiguration")
		createAddonsConfiguration(cl, "kubeless-test", "kubeless-addon-redis")
		ensureAddonsConfigurationIsReady(cl, "kubeless-test", "kubeless-addon-redis")
		log.Println("Deploying svc-instance")
		deployK8s("svc-instance.yaml")
		ensureSvcInstanceIsDeployed("kubeless-test", "redis")
		log.Println("Deploying service binding")
		deployK8s("svc-binding.yaml")
		ensureServceBindingIsReady("kubeless-test", "redis-binding")
		log.Println("Deploying svcbind-lambda")
		deployK8s("k8syaml/svcbind-lambda.yaml")
		ensureFunctionIsRunning("kubeless-test", "test-svcbind", true)
		log.Println("Verifying correct function output for test-svcbind")
		host := fmt.Sprintf("https://test-svcbind.%s", os.Getenv("DOMAIN_NAME"))
		ensureOutputIsCorrect(host, "OK", testID, "kubeless-test", "test-svcbind")
		log.Println("Verifying service connection for test-svcbind")
		ensureCorrectLog("kubeless-test", "test-svcbind", testDataRegex, testID, true)
		log.Println("Function test-svcbind works correctly")
	}()

	wg.Wait()
	cleanup(cl)
	log.Println("Success")
}

func createAddonsConfiguration(cl client.Client, ns string, name string) {
	ac := v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: "https://github.com/kyma-project/addons/releases/download/0.8.0/index-testing.yaml",
					},
				},
			},
		},
	}

	err := cl.Create(context.Background(), &ac)
	if err != nil {
		log.Fatalf("Error while creating the addons configuration")
	}
}

func ensureAddonsConfigurationIsReady(cl client.Client, ns string, name string) {
	key := client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			ac := v1alpha1.AddonsConfiguration{}
			err := cl.Get(context.Background(), key, &ac)
			if err != nil {
				log.Fatalf("Error fetching AddonsConfiguration %s: %v", key, err)
			}
			if ac.Status.Phase == v1alpha1.AddonsConfigurationReady {
				log.Printf("AddonsConfiguration has been successfully configured.")
				return
			}
			log.Fatalf("Timeout waiting to get AddonsConfiguration ready %s: %v", key, ac.Status)
		case <-tick:
			ac := v1alpha1.AddonsConfiguration{}
			err := cl.Get(context.Background(), key, &ac)
			if err != nil {
				log.Fatalf("Error fetching AddonsConfiguration %s: %v", key, err)
			}
			if ac.Status.Phase == v1alpha1.AddonsConfigurationReady {
				log.Printf("AddonsConfiguration has been successfully configured.")
				return
			}
		}
	}
}

func deleteAddonsConfiguration(cl client.Client, ns string, name string) {
	ac := v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	err := cl.Delete(context.Background(), &ac)
	if err != nil && !apierror.IsNotFound(err) {
		log.Fatalf("Error while deleting the AddonsConfiguration %s/%s: %v", ns, name, err)
	}
}
