package restore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

const functionYaml = "function.yaml"
const functionName = "hello"

var testUUID = uuid.New()
var namespace = "restore-test-" + testUUID.String()
var backupName = "test-" + testUUID.String()

func need(needed int, expected []interface{}) string {
	if len(expected) != needed {
		return fmt.Sprintf("Function needs %v arguments but got %v", needed, len(expected))
	}
	return ""
}

func shouldReturnSubstringEventually(actual interface{}, expected ...interface{}) string {
	if fail := need(3, expected); fail != "" {
		return fail
	}

	action, _ := actual.(func() string)

	if action == nil {
		return "the function needs to return a string \"func() string\""
	}

	until, _ := expected[1].(time.Duration)
	interval, _ := expected[2].(time.Duration)
	substring, _ := expected[0].(string)

	timeout := time.After(until)
	tick := time.Tick(interval)
	debugMessage := ""
	for {
		select {
		case <-timeout:
			debugMessage += "\n" + printLogsFunctionPodContainers(namespace, functionName)
			return fmt.Sprintf("Timeout: %v", debugMessage)

		case <-tick:
			value := action()
			debugMessage += "\n" + value
			if strings.Contains(value, substring) {
				return ""
			}
		}
	}
}

func TestBackupAndRestoreCluster(t *testing.T) {

	Convey("Setup", t, func() {
		Convey("Create namespace", func() {
			So(createNamespace(namespace), ShouldContainSubstring, "created")
			Convey("Create function", func() {
				createFunction(functionYaml, namespace)
				So(func() string {
					return getFunctionState(namespace, functionName)
				}, shouldReturnSubstringEventually, "Running", 60*time.Second, 1*time.Second)
				So(func() string {
					host := fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
					value, err := getFunctionOutput(host, namespace, functionName)
					if err != nil {
						t.Fatalf("Could not get function output for host%v: %v", host, err)
					}
					return value
				}, shouldReturnSubstringEventually, testUUID.String(), 60*time.Second, 1*time.Second)
			})
		})
	})

	Convey("Backup Cluster", t, func() {
		So(func() string {
			stdout, stderr, err := runCommand("ark", []string{"backup", "create", backupName, "--include-namespaces", namespace})
			if err != nil {
				t.Fatalf("Error was: %v \n %v", err, stderr.String())
			}
			return stdout.String()
		}(), ShouldContainSubstring, "submitted successfully")

		Convey("Check backup status", func() {
			So(func() string {
				stdout, stderr, err := runCommand("ark", []string{"backup", "get", backupName, "-oyaml"})
				if err != nil {
					t.Fatalf("Error was: %v \n %v", err, stderr.String())
				}
				return stdout.String()
			}, shouldReturnSubstringEventually, "phase: Completed", 60*time.Second, 1*time.Second)
			Convey("Delete resources from cluster", func() {
				_, stderr, _ := runCommand("kubectl", []string{"delete", "ns", namespace, "--grace-period=0", "--force", "--ignore-not-found"})
				So(stderr.String(), ShouldNotContainSubstring, "Error from server (Conflict): Operation cannot be fulfilled on namespaces")
				So(stderr.String(), ShouldNotContainSubstring, "No resources found")
				So(func() string {
					_, stderr, _ := runCommand("kubectl", []string{"get", "ns", namespace, "-oyaml"})
					return stderr.String()
				}, shouldReturnSubstringEventually, "NotFound", 1*time.Minute, 1*time.Second)
			})

		})

	})

	Convey("Restore Cluster", t, func() {
		So(func() string {
			stdout, stderr, err := runCommand("ark", []string{"restore", "create", "--from-backup", backupName, backupName})
			if err != nil {
				t.Fatalf("Error was: %v \n %v", err, stderr.String())
			}
			return stdout.String()
		}(), ShouldContainSubstring, "submitted successfully.")
		So(func() string {
			stdout, _, _ := runCommand("ark", []string{"restore", "get", backupName})
			return stdout.String()
		}, shouldReturnSubstringEventually, "Completed", 5*time.Minute, 1*time.Second)
		Convey("Test restored resources", func() {
			So(func() string {
				return getFunctionState(namespace, functionName)
			}, shouldReturnSubstringEventually, "Running", 60*time.Second, 1*time.Second)
			So(func() string {
				value, err := getFunctionOutput(fmt.Sprintf("http://%s.%s:8080", functionName, namespace), namespace, functionName)
				if err != nil {
					t.Fatalf("Could not get function output: %v", err)
				}
				return value
			}, shouldReturnSubstringEventually, testUUID.String(), 60*time.Second, 1*time.Second)

		})

	})

}

func runCommand(command string, args []string) (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout, stderr, err
}

func createNamespace(namespace string) string {
	command := "kubectl"
	args := []string{"create", "namespace", namespace}
	stdout, stderr, err := runCommand(command, args)
	if err != nil {
		return stderr.String()
	}
	return stdout.String()
}

func createFunction(functionYaml, namespace string) {
	command := "kubectl"
	args := []string{"apply", "-f", functionYaml, "-n", namespace}
	runCommand(command, args)
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

func getFunctionState(namespace, name string) (output string) {
	stdout, _, err := runCommand("kubectl", []string{"-n", namespace, "get", "pod", "-l", "function=" + name, "-ojsonpath={range .items[*]}{.status.phase}{end}"})

	if err != nil {
		//context.Fatalf("Error while fetching the status phase of the function pod when verifying function is running: %v", stdout.String(), stderr.String())
	}

	return stdout.String()
}

func printLogsFunctionPodContainers(namespace, name string) string {
	log.Fatalln(name, namespace)
	functionPodsCmd := exec.Command("kubectl", "-n", namespace, "get", "pod", "-l", "function="+name, "-ojsonpath={.items[0].metadata.name}")
	functionPodName, err := functionPodsCmd.CombinedOutput()
	var logs string
	if err != nil {
		logs += fmt.Sprintf("Error is fetch function pod: %v\n", string(functionPodName))
	}

	logs += fmt.Sprintf("---------- Logs from all containers for function pod: %s ----------\n", string(functionPodName))

	prepareContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), "prepare")

	prepareContainerLog, _ := prepareContainerLogCmd.CombinedOutput()
	logs += fmt.Sprintf("Logs from prepare container:\n%s\n", string(prepareContainerLog))

	installContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), "install")

	installContainerLog, _ := installContainerLogCmd.CombinedOutput()
	logs += fmt.Sprintf("Logs from prepare container:\n%s\n", string(installContainerLog))

	functionContainerLogCmd := exec.Command("kubectl", "-n", namespace, "logs", string(functionPodName), name)

	functionContainerLog, _ := functionContainerLogCmd.CombinedOutput()
	logs += fmt.Sprintf("Logs from %s container in pod %s:\n%s\n", name, string(functionPodName), string(functionContainerLog))

	envoyLogsCmd := exec.Command("kubectl", "-n", namespace, "log", "-l", string(functionPodName), "-c", "istio-proxy")

	envoyLogsCmdStdErr, _ := envoyLogsCmd.CombinedOutput()
	logs += fmt.Sprintf("Envoy Logs are:\n%s\n", string(envoyLogsCmdStdErr))
	return logs
}

func getFunctionOutput(host, namespace, name string) (string, error) {
	resp, err := http.Post(host, "text/plain", bytes.NewBuffer([]byte(testUUID.String())))
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "Unable to get response: %v", err
		}
		return string(bodyBytes), err
	}
	return "", err

}
