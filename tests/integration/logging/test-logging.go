package main

import (
	"log"
	"os/exec"

	"github.com/kyma-project/kyma/tests/integration/logging/pkg/fluentbit"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/jwt"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/logstream"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/loki"
)

const namespace = "kyma-system"
const yamlFile = "testCounterPod.yaml"

func main() {
	log.Println("Starting logging test")
	logstream.Cleanup(namespace)
	log.Println("Test if all the Loki/Fluent Bit pods are ready")
	loki.TestPodsAreReady()
	log.Println("Test if Fluent Bit is able to find Loki")
	fluentbit.Test()
	log.Println("Deploying test-counter-pod")
	deployDummyPod(namespace)
	log.Println("Test if logs from test-counter-pod are streamed by Loki")
	testLogStream(namespace)
	logstream.Cleanup(namespace)
}

func deployDummyPod(namespace string) {
	cmd := exec.Command("kubectl", "-n", namespace, "create", "-f", yamlFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Unable to deploy:\n", string(stdoutStderr))
	}
}

func testLogStream(namespace string) {
	logstream.WaitForDummyPodToRun(namespace)
	token := jwt.GetToken()
	authHeader := jwt.SetAuthHeader(token)
	logstream.Test("container", "count", authHeader, 0)
	logstream.Test("app", "test-counter-pod", authHeader, 0)
	logstream.Test("namespace", namespace, authHeader, 0)
	log.Println("Test Logging Succeeded!")
}
