package restore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

func createNamespace(namespace string) string {
	command := "kubectl"
	args := []string{"create", "namespace", namespace}
	stdout, stderr, err := runCommand(command, args)
	result := ""
	if err != nil {
		return stderr.String()
	}
	result = stdout.String()
	args = []string{"label", "namespace", namespace, "env=true", "istio-injection=enabled"}
	stdout, stderr, err = runCommand(command, args)
	if err != nil {
		return stderr.String()
	}
	result += stdout.String()
	return result
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

func getFunctionOutput(host, namespace, name, testUUID string) (string, error) {
	resp, err := http.Post(host, "text/plain", bytes.NewBufferString(testUUID))
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
