package main

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

var functionFile = "hello.js"
var ingressFile = "ingress.yaml"
var routeRuleFile = "route.yaml"
var dependencyFile = "dependecies.json"
var environment = "kubeless-test"

func podStatusIsRunning() bool {
	status := ""
	timeout := time.After(6 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", environment, "describe", "pods", "-l", "function=helloworld")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to describe the pods: %v", err)
			}
			log.Fatalf("Timed out waiting for Functions pods to be running: %v", string(stdoutStderr))
			return false
		case <-tick:
			cmd := exec.Command("kubectl", "-n", environment, "get", "pods", "-l", "function=helloworld", "--template={{range .items}}{{.status.phase}}{{end}}")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Unable to get the status: %v", err)
			}
			status = string(stdoutStderr)
			log.Println("Status is: ", status)
			if status == "Running" {
				return true
			}
		}
	}
}

func checkCurlOutput() (string, bool) {
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			log.Printf("Timed out waiting for curl")
			cmd := exec.Command("curl", "-s", "http://istio-ingress.istio-system/hello")
			stdoutStderr, _ := cmd.CombinedOutput()
			return string(stdoutStderr), false
		case <-tick:
			cmd := exec.Command("curl", "-s", "http://istio-ingress.istio-system/hello")
			stdoutStderr, _ := cmd.CombinedOutput()
			if string(stdoutStderr) == "hello world" {
				return string(stdoutStderr), true
			}
		}
	}
}

func checkPodsDeleted() bool {
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			log.Printf("Timed out waiting for pods from previous deployment to be deleted")
			cmd := exec.Command("kubectl", "get", "po", "-l", "function=helloworld", "--template={{range .items}}{{.status.phase}}{{end}}")
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatalf("Unable to delete pods from previous deployment: %v", stdoutStderr)

		case <-tick:
			cmd := exec.Command("kubectl", "-n", environment, "get", "po", "-l", "function=helloworld", "--template={{range .items}}{{.status.phase}}{{end}}")
			stdoutStderr, _ := cmd.CombinedOutput()
			if len(string(stdoutStderr)) == 0 || strings.Contains(string(stdoutStderr), "NotFound") {
				return true
			}
		}
	}
}

func checkNamespaceDeleted() bool {
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			log.Printf("Timed out waiting for namespace to be deleted")
			cmd := exec.Command("kubectl", "get", "ns", environment)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatalf("Unable to delete namespace from previous test run: %v", stdoutStderr)

		case <-tick:
			cmd := exec.Command("kubectl", "get", "ns", environment)
			stdoutStderr, _ := cmd.CombinedOutput()
			if len(string(stdoutStderr)) != 0 && strings.Contains(string(stdoutStderr), "NotFound") {
				return true
			}
		}
	}

}

func cleanupPlayground() bool {
	cmd := exec.Command("kubeless", "function", "delete", "helloworld", "--namespace", environment)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(stdoutStderr), "not found") {
		log.Fatalf("Unable to delete previous deployment of function 'helloworld' : %v", string(stdoutStderr))
	}

	cmd = exec.Command("kubectl", "-n", environment, "delete", "po", "-l", "function=helloworld", "--grace-period=0", "--force")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to delete pods of deployment of function 'helloworld' : %v", string(stdoutStderr))
	}
	if checkPodsDeleted() {
		log.Println("Successfully deleted pods from previous deployment")
	}

	cmd = exec.Command("kubectl", "-n", environment, "delete", "-f", ingressFile)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(stdoutStderr), "NotFound") {
		log.Fatalf("Unable to delete ingress: %v", string(stdoutStderr))
	}
	log.Println("Deleted ingress: ", string(stdoutStderr))

	cmd = exec.Command("kubectl", "-n", environment, "delete", "-f", routeRuleFile)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(stdoutStderr), "NotFound") {
		log.Fatalf("Unable to delete Route rule: %v", string(stdoutStderr))
	}
	log.Println("Deleted Route rule: ", string(stdoutStderr))

	cmd = exec.Command("kubectl", "delete", "ns", environment)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(stdoutStderr), "not found") {
		log.Fatalf("Unable to delete namespace 'kubeless-test' : %v", string(stdoutStderr))
	}
	if checkNamespaceDeleted() {
		log.Println("Successfully deleted namespace 'kubeless-test' from previous test run")
	}

	return true
}

func main() {
	log.Println("Starting Test")

	if cleanupPlayground() == false {
		log.Fatal("Previous deployment of function 'helloworld' has not been cleanedup. Bailing out!!")
	}

	log.Printf("Creating a new environment: %v", environment)
	cmd := exec.Command("kubectl", "create", "ns", environment, "label", "env=true")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to create environment to test : %v", string(stdoutStderr))
	}

	log.Printf("Applying 'env' label to the new environment: %v", environment)
	cmd = exec.Command("kubectl", "label", "ns", environment, "env=true")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to label environment : %v", string(stdoutStderr))
	}

	cmd = exec.Command("kubeless", "function", "deploy", "helloworld", "--namespace", environment, "--runtime", "nodejs6", "--from-file", functionFile, "--handler", "hello.foo", "--dependencies", dependencyFile)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to deploy function via kubeless: %v", string(stdoutStderr))
	}
	log.Println("Deployed function: ", string(stdoutStderr))

	if podStatusIsRunning() == false {
		log.Fatal("The function pods are not running")
	}

	cmd = exec.Command("kubectl", "apply", "-n", environment, "-f", ingressFile)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to deploy ingress: %v", string(stdoutStderr))
	}
	log.Println("Deployed ingress: ", string(stdoutStderr))

	cmd = exec.Command("kubectl", "-n", environment, "apply", "-f", routeRuleFile)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to deploy Route rule: %v", string(stdoutStderr))
	}
	log.Println("Deployed Route rule: ", string(stdoutStderr))

	resp, pass := checkCurlOutput()

	if pass {
		log.Printf("Got expected output 'hello world'=='%v'", resp)
	} else {
		log.Fatalf("Did not get the expected output 'hello world' but got: %v", resp)
	}

	if cleanupPlayground() == false {
		log.Fatal("Previous deployment of function 'hellowrld' has not been cleanedup. Bailing out!!")
	}

	log.Println("Successfully finished the test")

}
