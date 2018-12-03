// hpa-functions_test.go
package main

import (
	"runtime"
	"net/http"
	"log"
	"io/ioutil"
	"crypto/tls"
	"sync"
	"time"
	"fmt"
	"os/exec"
	"os"
	"strings"
)

const (

	numRequest = 200
)

var (
	url = "test-hpa-test.kyma.local/"
)

func init() {
	cleanup()
	// time.Sleep(10 * time.Second)

	log.Println("Starting test")
	log.Printf("Domain Name is: %v", os.Getenv("DOMAIN_NAME"))

	runtime.GOMAXPROCS(runtime.NumCPU())

	deployK8s("k8syaml/ns.yaml")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Println("Deploying test-autoscaler function")
		deployFun("hpa-test", "test-autoscaler", "nodejs6", "js/test.js", "test.handler")
		log.Println("Verifying correct function output for test-autoscaler")
		url = fmt.Sprintf("https://test-autoscaler.%s", os.Getenv("DOMAIN_NAME"))
		//ensureOutputIsCorrect(host, "hello world", testID, "kubeless-test", "test-hello")
		log.Printf("Function test-autoscaler has been created url: %s \n", url)
		defer wg.Done()
	}()

}


func deployFun(namespace, name, runtime, codeFile, handler string) {
	cmd := exec.Command("kubeless", "-n", namespace, "function", "deploy", name, "-r", runtime, "-f", codeFile, "--handler", handler)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy function ", name, ":\n", string(stdoutStderr))
	}
	ensureFunctionIsRunning(namespace, name, false)
}


func ensureFunctionIsRunning(namespace, name string, serviceBinding bool) {
	timeout := time.After(6 * time.Minute)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			cmd := exec.Command("kubectl", "-n", namespace, "describe", "pod", "-l", "function="+name)
			stdoutStderr, _ := cmd.CombinedOutput()
			log.Fatalf("[%v] Timed out waiting for: %v function pod to be running. Because of following error: %v ", name, name, string(stdoutStderr))
		case <-tick:
			cmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={range .items[*]}{.status.phase}{end}")
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
				log.Printf("[%v] Pod: %v: ", name, string(functionPodName))
				return
			}
		}
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


func deployK8s(yamlFile string) {
	cmd := exec.Command("kubectl", "create", "-f", yamlFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy:\n", string(stdoutStderr))
	}
}

func main() {

	numGoroutines := numRequest * runtime.GOMAXPROCS(runtime.NumCPU())
	log.Println("logical CPUs: ", runtime.NumCPU())
	log.Println("number of goroutines: ", numGoroutines)

	client := getHttpClient(true)
	start := time.Now()
	ch := make(chan string)
	doneCh := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func() {
			defer wg.Done()
			//makeHttpRequest(client, ch)
			callFunction(client, ch)
		}()
	}

	go func() {
		printResponse(ch, doneCh)
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	<-doneCh

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}


func printResponse(ch chan string, doneCh chan bool) {
	for resp := range ch {
		log.Println(resp)
	}
	doneCh <- true
	log.Println("Done!")
}


func makeHttpRequest(client *http.Client, ch chan<-string) {
	start := time.Now()

	// request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error .\n[ERRO] -", err)
	}

	// GET request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}
	defer resp.Body.Close()

	secs := time.Since(start).Seconds()
	body, _ := ioutil.ReadAll(resp.Body)
	ch <- fmt.Sprintf("%.2f elapsed with response: %s %s", secs, string([]byte(body)), url)
}


func callFunction(client *http.Client, ch chan<-string) {
	timeout := time.After(5 * time.Second)
	call := time.Tick(10 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return
		case <-call:
			makeHttpRequest(client, ch)
		}
	}

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
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		deleteK8s("k8syaml/api.yaml")
		defer wg.Done()
	}()
	//go func() {
	//	deleteK8s("k8syaml/function.yaml")
	//	defer wg.Done()
	//}()
	go func() {
		deleteK8s("k8syaml/ns.yaml")
		defer wg.Done()
	}()
	wg.Wait()
}
