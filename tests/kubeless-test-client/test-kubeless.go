package main

import (
	"log"
	"math/rand"
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

func ensureOutputIsCorrect(url, host, expectedOutput, testID string) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("curl", "-s", url, "-H", "Host: "+host)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatal("Timed out getting the correct output from ", url, " host ", host, ":\n", string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("curl", "-ks", url, "-H", "Host: "+host, "-d", testID)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatal("Unable to call ", url, " host ", host, ":\n", string(stdoutStderr))
			}
			if string(stdoutStderr) == expectedOutput {
				return
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

func ensureCorrectLog(namespace, funName string, pattern *regexp.Regexp, match string) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "log", "-l", "function="+funName, "-c", funName)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatal("Timed out getting the correct log for ", funName, ":\n", string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "log", "-l", "function="+funName, "-c", funName)
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

func cleanup() {
	log.Println("Cleaning up")
	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		deleteK8s("k8s.yaml")
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
		deleteK8s("svcbind.yaml")
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
	testID := randomString(8)
	deployK8s("ns.yaml")
	deployK8s("k8s.yaml")
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		log.Println("Deploying test-hello function")
		deployFun("kubeless-test", "test-hello", "nodejs6", "dependencies.json", "hello.js", "hello.handler")
		log.Println("Verifying correct function output for test-hello")
		ensureOutputIsCorrect("https://istio-ingress.istio-system", "test-hello.kyma.local", "hello world", testID)
		log.Println("Function test-hello works correctly")
		defer wg.Done()
	}()

	go func() {
		log.Println("Deploying test-event function")
		deployFun("kubeless-test", "test-event", "nodejs6", "dependencies.json", "event.js", "event.handler")
		log.Println("Publishing event to function test-event")
		publishEvent(testID)
		log.Println("Verifying correct event processing for test-event")
		ensureCorrectLog("kubeless-test", "test-event", testDataRegex, testID)
		log.Println("Function test-event works correctly")
		defer wg.Done()
	}()

	go func() {
		log.Println("Deploying test-svcbind function")
		deployK8s("svcbind.yaml")
		ensureFunctionIsRunning("kubeless-test", "test-svcbind", true)
		log.Println("Verifying correct function output for test-svcbind")
		ensureOutputIsCorrect("https://istio-ingress.istio-system", "test-svcbind.kyma.local", "OK", testID)
		log.Println("Verifying service connection for test-svcbind")
		ensureCorrectLog("kubeless-test", "test-svcbind", testDataRegex, testID)
		log.Println("Function test-svcbind works correctly")
		defer wg.Done()
	}()

	wg.Wait()
	cleanup()
	log.Println("Success")
}
