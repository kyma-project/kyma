package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
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
	cmd := exec.Command("kubectl", "delete", "-f", yamlFile)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "NotFound") {
		log.Fatal("Unable to delete:\n", output)
	}
}

func deleteNamespace(namespace string) {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			log.Fatal("Timed out waiting for namespace ", namespace, " to be deleted\n")
		case <-tick:
			cmd := exec.Command("kubectl", "delete", "ns", namespace)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil && strings.Contains(string(stdoutStderr), "NotFound") {
				return
			}
		}
	}
}

func deployFun(namespace, name, runtime, depFile, codeFile, handler string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "deploy", name, "-r", runtime, "-d", depFile, "-f", codeFile, "--handler", handler)
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
	}

	timeout := time.After(6 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "describe", "pod", "-l", "function="+name)
			if sbuID != "" {
				cmd = exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-l", "use-"+sbuID)
			}
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatal("Timed out waiting for ", name, " pod to be running:\n", string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={range .items[*]}{.status.phase}{end}")
			if sbuID != "" {
				cmd = exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-l", "use-"+sbuID, "-ojsonpath={range .items[*]}{.status.phase}{end}")
			}
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(stdoutStderr), "Running") {
				return
			}
		}
	}
}

func deleteFun(namespace, name string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "delete", name)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil && !strings.Contains(output, "not found") {
		log.Fatal("Unable to delete function ", name, ":\n", output)
	}

	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			log.Fatal("Timed out waiting for ", name, " pod to be deleted\n")
		case <-tick:
			cmd = exec.Command("kubectl", "-n", namespace, "delete", "pod", "-l", "function="+name)
			stdoutStderr, err := cmd.CombinedOutput()
			if err == nil && strings.Contains(string(stdoutStderr), "No resources found") {
				return
			}
		}
	}
}

func getMinikubeIP() string {
	mipCmd := exec.Command("minikube", "ip")
	if mipOut, err := mipCmd.Output(); err != nil {
		log.Fatalf("Error while getting minikube IP. Root cause: %s", err)
		return ""
	} else {
		return strings.Trim(string(mipOut), "\n")
	}
}

func ensureOutputIsCorrect(host, expectedOutput, testID, namespace, testName string) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)

	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	ingressGatewayControllerServiceURL := "istio-ingressgateway.istio-system.svc.cluster.local"

	ingressGatewayControllerAddr, err := net.LookupHost(ingressGatewayControllerServiceURL)
	if err != nil {
		log.Printf("Unable to resolve host '%s'. Root cause: %v", ingressGatewayControllerServiceURL, err)
		if minikubeIP := getMinikubeIP(); minikubeIP != "" {
			ingressGatewayControllerAddr = []string{minikubeIP}
		}
	}
	if len(ingressGatewayControllerAddr) > 0 {
		log.Printf("Ingress controller address: '%s'", ingressGatewayControllerAddr[0])

		http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			addr = ingressGatewayControllerAddr[0] + ":443"
			return dialer.DialContext(ctx, network, addr)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		for {
			select {
			case <-timeout:
				log.Println("Timeout: Check if virtual service has been created")
				cmd := exec.Command("kubectl", "-n", namespace, "get", "virtualservices.networking.istio.io")
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatal("Unable to fetch virtual Service for", testName, ":\n", err)
				}
				log.Printf("Timeout: Virtual Service is: %v", string(stdoutStderr))

				log.Println("Timeout: Check http Get one last time")
				resp, err := http.Post(host, "text/plain", bytes.NewBuffer([]byte(testID)))
				if err != nil {
					log.Fatal("Timeout: Unable to call host ", host, ":\n", err)
				}
				log.Fatalf("Timeout: Response is: %v", resp)

			case <-tick:
				resp, err := http.Post(host, "text/plain", bytes.NewBuffer([]byte(testID)))
				if err != nil {
					log.Fatal("Unable to call host ", host, ":\n", err)
				}
				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatalf("Unable to get response: %v", err)
					}
					if string(bodyBytes) == expectedOutput {
						log.Printf("Response is equal to expected output: %v == %v", string(bodyBytes), expectedOutput)
						return
					}
					log.Fatalf("Response is not equal to expected output: %v != %v", string(bodyBytes), expectedOutput)
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

func publishEvent(testID string) {
	cmd := exec.Command("curl", "-s", "http://core-publish:8080/v1/events", "-H", "Content-Type: application/json", "-d", `{"source": {"source-namespace": "test", "source-type": "test", "source-environment": "test"}, "event-type": "test", "event-type-version": "v1", "event-time": "0001-01-01T00:00:00+00:00", "data": "`+testID+`"}`)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to publish event:\n", string(stdoutStderr))
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
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "serviceinstance", svcInstance, "--output=jsonpath={.items[0].metadata.name}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to fetch service instance %v: %v", svcInstance, err)
			}
			log.Fatalf("Timeout waiting to get service instance %v: %v", svcInstance, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "serviceinstance", svcInstance, "-o=jsonpath={.items[*]}{.status.conditions[*].reason}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error fetching service instance %v: %v", svcInstance, err)
			}
			if string(stdoutStderr) == "ProvisionedSuccessfully" {
				return
			}
		}
	}
}

func cleanup() {
	log.Println("Cleaning up")
	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		deleteK8s("k8syaml/k8s.yaml")
		defer wg.Done()
	}()
	go func() {
		deleteFun("kubeless-test", "test-hello")
		defer wg.Done()
	}()
	go func() {
		deleteFun("kubeless-test", "test-event")
		defer wg.Done()
	}()
	go func() {
		deleteK8s("k8syaml/svcbind-lambda.yaml")
		deleteK8s("svc-instance.yaml")
		defer wg.Done()
	}()
	go func() {
		deleteNamespace("kubeless-test")
		defer wg.Done()
	}()
	wg.Wait()
}

var testDataRegex = regexp.MustCompile(`(?m)^OK ([a-z0-9]{8})$`)

func main() {
	cleanup()
	rand.Seed(time.Now().UTC().UnixNano())

	log.Println("Starting test")
	log.Printf("Domain Name is: %v", os.Getenv("DOMAIN_NAME"))
	testID := randomString(8)
	deployK8s("ns.yaml")
	deployK8s("k8syaml/k8s.yaml")
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		log.Println("Deploying test-hello function")
		deployFun("kubeless-test", "test-hello", "nodejs6", "dependencies.json", "hello.js", "hello.handler")
		log.Println("Verifying correct function output for test-hello")
		host := fmt.Sprintf("https://test-hello.%s", os.Getenv("DOMAIN_NAME"))
		ensureOutputIsCorrect(host, "hello world", testID, "kubeless-test", "test-hello")
		log.Println("Function test-hello works correctly")
		defer wg.Done()
	}()

	go func() {
		log.Println("Deploying test-event function")
		deployFun("kubeless-test", "test-event", "nodejs6", "dependencies.json", "event.js", "event.handler")
		log.Println("Publishing event to function test-event")
		publishEvent(testID)
		log.Println("Verifying correct event processing for test-event")
		ensureCorrectLog("kubeless-test", "test-event", testDataRegex, testID, false)
		log.Println("Function test-event works correctly")
		defer wg.Done()
	}()

	go func() {
		log.Println("Deploying svc-instance")
		deployK8s("svc-instance.yaml")
		ensureSvcInstanceIsDeployed("kubeless-test", "redis")
		log.Println("Deploying svcbind-lambda")
		deployK8s("k8syaml/svcbind-lambda.yaml")
		ensureFunctionIsRunning("kubeless-test", "test-svcbind", true)
		log.Println("Verifying correct function output for test-svcbind")
		host := fmt.Sprintf("https://test-svcbind.%s", os.Getenv("DOMAIN_NAME"))
		ensureOutputIsCorrect(host, "OK", testID, "kubeless-test", "test-svcbind")
		log.Println("Verifying service connection for test-svcbind")
		ensureCorrectLog("kubeless-test", "test-svcbind", testDataRegex, testID, true)
		log.Println("Function test-svcbind works correctly")
		defer wg.Done()
	}()

	wg.Wait()
	cleanup()
	log.Println("Success")
}
